package orms

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// validateDBName ensures dbName is a safe, bare SQL identifier so it cannot
// be used for SQL injection in interpolated DDL.
func validateDBName(dbName string) error {
	for i, r := range dbName {
		valid := r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (i > 0 && r >= '0' && r <= '9')
		if !valid {
			return fmt.Errorf("invalid database name %q: must be a bare identifier (letters, digits, underscore)", dbName)
		}
	}
	if dbName == "" {
		return fmt.Errorf("invalid database name: must not be empty")
	}
	return nil
}

// schemaOf parses the GORM schema for a model value and returns it.
func schemaOf(tx *gorm.DB, model any) (*schema.Schema, error) {
	stmt := &gorm.Statement{DB: tx}
	if err := stmt.Parse(model); err != nil {
		return nil, err
	}
	return stmt.Schema, nil
}

// desiredColumns returns the struct-declared column names that participate in
// migration, in declaration order (embedded base models are flattened already).
func desiredColumns(s *schema.Schema) []string {
	var names []string
	for _, name := range s.DBNames {
		if field, ok := s.FieldsByDBName[name]; ok && field.IgnoreMigration {
			continue
		}
		names = append(names, name)
	}
	return names
}

// reorderColumns moves every column so the physical order matches the given
// order (struct declaration first, soft-mode extras last). MySQL/MariaDB
// reorders in place; other drivers fall back to a rebuild that copies shared
// columns.
func reorderColumns(tx *gorm.DB, model any, s *schema.Schema, order []string) error {
	switch tx.Name() {
	case "mysql", "mariadb":
		return reorderInPlace(tx, model, s, order)
	default:
		return rebuildTable(tx, model, s, order)
	}
}

// reorderInPlace rewrites each out-of-position column with FIRST/AFTER, which on
// MySQL also normalizes the column type/comment back to the struct declaration.
// Columns without a struct field (soft-mode extras) keep their type as
// reconstructed from the database.
func reorderInPlace(tx *gorm.DB, model any, s *schema.Schema, order []string) error {
	var prev string
	for i, name := range order {
		pos, err := columnPositions(tx, model)
		if err != nil {
			return err
		}
		if _, ok := pos[name]; !ok {
			prev = name
			continue
		}
		if pos[name] == i {
			prev = name
			continue
		}
		field, _ := s.FieldsByDBName[name]
		var fullType any
		if field == nil {
			// soft-mode extra column: reconstruct its definition from the database
			columnType, err := columnType(tx, model, name)
			if err != nil {
				prev = name
				continue
			}
			ddl, err := columnDefSQL(columnType)
			if err != nil {
				prev = name
				continue
			}
			fullType = clause.Expr{SQL: ddl}
		} else {
			fullType = tx.Migrator().FullDataTypeOf(field)
		}
		var sql string
		var values []any
		if i == 0 {
			sql = "ALTER TABLE ? MODIFY COLUMN ? ? FIRST"
			values = []any{clause.Table{Name: s.Table}, clause.Column{Name: name}, fullType}
		} else {
			sql = "ALTER TABLE ? MODIFY COLUMN ? ? AFTER ?"
			values = []any{clause.Table{Name: s.Table}, clause.Column{Name: name}, fullType, clause.Column{Name: prev}}
		}
		fmt.Printf("sync %s: reorder column %s\n", s.Table, name)
		if err := tx.Exec(sql, values...).Error; err != nil {
			return err
		}
		prev = name
	}
	return nil
}

// rebuildTable recreates a table in struct order and copies shared columns back,
// used for drivers that cannot reorder columns in place. Soft-mode extras that
// survive are re-added afterwards with their type reconstructed from the
// database so they keep their data.
func rebuildTable(tx *gorm.DB, model any, s *schema.Schema, order []string) error {
	types, err := tx.Migrator().ColumnTypes(model)
	if err != nil {
		return err
	}
	actual := map[string]gorm.ColumnType{}
	for _, col := range types {
		name := col.Name()
		actual[name] = col
	}

	tmp := s.Table + "_sync_tmp"
	if err := tx.Exec("DROP TABLE IF EXISTS ?", clause.Table{Name: tmp}).Error; err != nil {
		return err
	}
	if err := tx.Migrator().RenameTable(s.Table, tmp); err != nil {
		return err
	}
	if err := AutoMigrate(tx, model); err != nil {
		return err
	}

	// re-add soft-mode extras that survived, in their target order
	for _, name := range order {
		if _, ok := s.FieldsByDBName[name]; ok {
			continue
		}
		col, ok := actual[name]
		if !ok {
			continue
		}
		ddl, err := columnDefSQL(col)
		if err != nil {
			return err
		}
		fmt.Printf("sync %s: add column %s\n", s.Table, name)
		if err := tx.Exec("ALTER TABLE ? ADD COLUMN ? ?", clause.Table{Name: s.Table}, clause.Column{Name: name}, clause.Expr{SQL: ddl}).Error; err != nil {
			return err
		}
	}

	common := []string{}
	for _, name := range order {
		if _, ok := actual[name]; ok {
			common = append(common, name)
		}
	}
	if len(common) > 0 {
		var sql strings.Builder
		sql.WriteString("INSERT INTO ? (")
		values := []any{clause.Table{Name: s.Table}}
		for i, name := range common {
			if i > 0 {
				sql.WriteString(",")
			}
			sql.WriteString("?")
			values = append(values, clause.Column{Name: name})
		}
		sql.WriteString(") SELECT ")
		for i, name := range common {
			if i > 0 {
				sql.WriteString(",")
			}
			sql.WriteString("?")
			values = append(values, clause.Column{Name: name})
		}
		sql.WriteString(" FROM ?")
		values = append(values, clause.Table{Name: tmp})
		if err := tx.Exec(sql.String(), values...).Error; err != nil {
			return err
		}
	}

	if err := tx.Exec("DROP TABLE IF EXISTS ?", clause.Table{Name: tmp}).Error; err != nil {
		return err
	}
	fmt.Printf("sync %s: rebuilt columns (%d)\n", s.Table, len(common))
	return nil
}

// columnPositions returns the actual column names mapped to their current
// ordinal position in the database.
func columnPositions(tx *gorm.DB, model any) (map[string]int, error) {
	types, err := tx.Migrator().ColumnTypes(model)
	if err != nil {
		return nil, err
	}
	pos := map[string]int{}
	for i, col := range types {
		pos[col.Name()] = i
	}
	return pos, nil
}

// columnType returns the database column type metadata for the named column.
func columnType(tx *gorm.DB, model any, name string) (gorm.ColumnType, error) {
	types, err := tx.Migrator().ColumnTypes(model)
	if err != nil {
		return nil, err
	}
	for _, col := range types {
		if col.Name() == name {
			return col, nil
		}
	}
	return nil, fmt.Errorf("column %s not found", name)
}

// columnDefSQL reconstructs a best-effort DDL definition for a database column
// from its runtime type metadata. The driver-provided full type (for example
// "varchar(64)") is preferred; length and precision are appended only when it is
// not available.
func columnDefSQL(col gorm.ColumnType) (string, error) {
	if full, ok := col.ColumnType(); ok && full != "" {
		var b strings.Builder
		b.WriteString(full)
		if nullable, ok := col.Nullable(); ok && !nullable {
			b.WriteString(" NOT NULL")
		}
		return b.String(), nil
	}
	name := col.DatabaseTypeName()
	if name == "" {
		return "", fmt.Errorf("unknown database type")
	}
	var b strings.Builder
	b.WriteString(name)
	if length, ok := col.Length(); ok && length > 0 && !strings.Contains(name, "(") {
		fmt.Fprintf(&b, "(%d)", length)
	} else if precision, scale, ok := col.DecimalSize(); ok && !strings.Contains(name, "(") {
		if scale > 0 {
			fmt.Fprintf(&b, "(%d,%d)", precision, scale)
		} else {
			fmt.Fprintf(&b, "(%d)", precision)
		}
	}
	nullable, _ := col.Nullable()
	if !nullable {
		b.WriteString(" NOT NULL")
	}
	return b.String(), nil
}
