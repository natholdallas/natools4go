package orms

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// columnOrder returns the physical column names of a model's table in ordinal
// order.
func columnOrder(tx *gorm.DB, model any) ([]string, error) {
	pos, err := columnPositions(tx, model)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(pos))
	for name := range pos {
		names = append(names, name)
	}
	slices.SortFunc(names, func(a, b string) int { return pos[a] - pos[b] })
	return names, nil
}

// commonColumns filters order down to the columns that exist in actual,
// preserving the requested order.
func commonColumns(actual, order []string) []string {
	present := make(map[string]bool, len(actual))
	for _, name := range actual {
		present[name] = true
	}
	cols := make([]string, 0, len(order))
	for _, name := range order {
		if present[name] {
			cols = append(cols, name)
		}
	}
	return cols
}

// quotedColumnList renders a comma-separated identifier list for a driver.
func quotedColumnList(driver string, cols []string) string {
	parts := make([]string, len(cols))
	for i, name := range cols {
		parts[i] = quoteIdent(driver, name)
	}
	return strings.Join(parts, ", ")
}

// countRows returns the row count of an already-quoted table reference.
func countRows(tx *gorm.DB, quotedTable string) (int64, error) {
	var n int64
	if err := tx.Raw("SELECT count(*) FROM " + quotedTable).Scan(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}

// reorderColumns moves the physical columns of table into order. Drivers that
// can reorder in place do so; the rest rebuild the table and copy the data.
func reorderColumns(tx *gorm.DB, model any, s *schema.Schema, order []string) error {
	actual, err := columnOrder(tx, model)
	if err != nil {
		return err
	}
	if slices.Equal(actual, order) {
		return nil
	}
	switch tx.Name() {
	case "mysql", "mariadb":
		return reorderMySQL(tx, s.Table, actual, order)
	case "sqlite", "sqlite3":
		return rebuildSQLite(tx, s.Table, actual, order)
	case "clickhouse":
		return rebuildClickHouse(tx, s.Table, actual, order)
	case "postgres", "postgresql":
		return rebuildPostgres(tx, s.Table, actual, order)
	case "sqlserver", "mssql":
		return rebuildSQLServer(tx, s.Table, actual, order)
	default:
		return fmt.Errorf("orms: column reorder not supported for driver %q", tx.Name())
	}
}

// columnDefs maps column name to the database's own DDL text for it, taken from
// a parsed CREATE TABLE item list.
func columnDefs(doc ddlDoc, columns []string) map[string]string {
	known := make(map[string]bool, len(columns))
	for _, name := range columns {
		known[name] = true
	}
	defs := make(map[string]string, len(columns))
	for _, item := range doc.items {
		if name, ok := itemIdent(item); ok && known[name] {
			defs[name] = item
		}
	}
	return defs
}

// reorderMySQL rewrites out-of-position columns with a single
// ALTER TABLE ... MODIFY COLUMN ... FIRST/AFTER. Each MODIFY reuses the column
// definition returned by SHOW CREATE TABLE, so DEFAULT, COMMENT, COLLATE,
// generated expressions and AUTO_INCREMENT survive unchanged. That definition
// text comes from the server, never from user input.
func reorderMySQL(tx *gorm.DB, table string, actual, order []string) error {
	var tableName, ddl string
	if err := tx.Raw("SHOW CREATE TABLE ?", clause.Table{Name: table}).Row().Scan(&tableName, &ddl); err != nil {
		return fmt.Errorf("orms: read definition of %s: %w", table, err)
	}
	doc, err := parseCreateTable(ddl)
	if err != nil {
		return err
	}
	defs := columnDefs(doc, actual)

	cur := slices.Clone(actual)
	var clauses []string
	var args []any
	for i, name := range order {
		at := slices.Index(cur, name)
		if at < 0 || at == i {
			continue
		}
		def, ok := defs[name]
		if !ok {
			return fmt.Errorf("orms: column %s not found in definition of %s", name, table)
		}
		modify := "MODIFY COLUMN " + def
		if i == 0 {
			modify += " FIRST"
		} else {
			modify += " AFTER ?"
			args = append(args, clause.Column{Name: order[i-1]})
		}
		clauses = append(clauses, modify)
		cur = slices.Insert(slices.Delete(cur, at, at+1), i, name)
	}
	if len(clauses) == 0 {
		return nil
	}
	query := "ALTER TABLE ? " + strings.Join(clauses, ", ")
	values := append([]any{clause.Table{Name: table}}, args...)
	return tx.Exec(query, values...).Error
}

// rebuildSQLite recreates the table from its own CREATE TABLE statement with
// the columns in the requested order, then copies the data.
//
// Everything runs on one dedicated connection: PRAGMA foreign_keys and temp
// table visibility are connection/session scoped. Foreign keys are disabled for
// the copy so dropping the old table neither cascades nor rewrites references
// in other tables; a foreign_key_check validates the result before commit.
func rebuildSQLite(tx *gorm.DB, table string, actual, order []string) error {
	return tx.Connection(func(db *gorm.DB) error {
		var fk int
		if err := db.Raw("PRAGMA foreign_keys").Scan(&fk).Error; err != nil {
			return err
		}
		if fk == 1 {
			if err := db.Exec("PRAGMA foreign_keys=OFF").Error; err != nil {
				return err
			}
			defer db.Exec("PRAGMA foreign_keys=ON")
		}

		type schemaObj struct {
			Type string
			Name string
			SQL  string
		}
		var objects []schemaObj
		if err := db.Raw(
			"SELECT type, name, sql FROM sqlite_master WHERE tbl_name = ? AND type IN ('index','trigger') AND sql IS NOT NULL",
			table,
		).Scan(&objects).Error; err != nil {
			return err
		}

		var ddl string
		if err := db.Raw("SELECT sql FROM sqlite_master WHERE type='table' AND name = ?", table).Row().Scan(&ddl); err != nil {
			return fmt.Errorf("orms: read definition of %s: %w", table, err)
		}
		doc, err := parseCreateTable(ddl)
		if err != nil {
			return err
		}
		reordered := reorderItems(doc, actual, order)

		newTable := table + "__sync_new"
		qold := quoteIdent("sqlite", table)
		qnew := quoteIdent("sqlite", newTable)
		createSQL := "CREATE TABLE " + qnew + " (" + strings.Join(reordered.items, ", ") + ")" + reordered.suffix

		cols := commonColumns(actual, order)
		colList := quotedColumnList("sqlite", cols)
		insertSQL := "INSERT INTO " + qnew + " (" + colList + ") SELECT " + colList + " FROM " + qold

		err = db.Transaction(func(tx2 *gorm.DB) error {
			if err := tx2.Exec("DROP TABLE IF EXISTS " + qnew).Error; err != nil {
				return err
			}
			if err := tx2.Exec(createSQL).Error; err != nil {
				return err
			}
			if err := tx2.Exec(insertSQL).Error; err != nil {
				return err
			}
			before, err := countRows(tx2, qold)
			if err != nil {
				return err
			}
			after, err := countRows(tx2, qnew)
			if err != nil {
				return err
			}
			if before != after {
				return fmt.Errorf("orms: refusing to rebuild %s: copied %d of %d rows", table, after, before)
			}
			if err := tx2.Exec("DROP TABLE " + qold).Error; err != nil {
				return err
			}
			if err := tx2.Exec("ALTER TABLE " + qnew + " RENAME TO " + qold).Error; err != nil {
				return err
			}
			for _, obj := range objects {
				if err := tx2.Exec(obj.SQL).Error; err != nil {
					return fmt.Errorf("orms: restore %s %s: %w", obj.Type, obj.Name, err)
				}
			}
			rows, err := tx2.Raw("PRAGMA foreign_key_check(" + qold + ")").Rows()
			if err != nil {
				return err
			}
			violated := rows.Next()
			rows.Close()
			if violated {
				return fmt.Errorf("orms: foreign key violation after rebuilding %s", table)
			}
			return nil
		})
		// A failed transaction leaves the temporary table behind; the original
		// table is still intact because it is only dropped inside the commit.
		if err != nil {
			db.Exec("DROP TABLE IF EXISTS " + qnew)
			return err
		}
		return nil
	})
}

// rebuildClickHouse recreates the table and swaps it in with a single atomic
// RENAME TABLE. ClickHouse has no transactions and no UPDATE/DELETE, so the row
// count is verified before the rename; any failure leaves the original table.
func rebuildClickHouse(tx *gorm.DB, table string, actual, order []string) error {
	var ddl string
	if err := tx.Raw("SHOW CREATE TABLE ?", clause.Table{Name: table}).Row().Scan(&ddl); err != nil {
		return fmt.Errorf("orms: read definition of %s: %w", table, err)
	}
	doc, err := parseCreateTable(ddl)
	if err != nil {
		return err
	}
	reordered := reorderItems(doc, actual, order)

	newTable := table + "__sync_new"
	oldTable := table + "__sync_old"
	qold := quoteIdent("clickhouse", table)
	qnew := quoteIdent("clickhouse", newTable)
	qoldBackup := quoteIdent("clickhouse", oldTable)

	cols := commonColumns(actual, order)
	colList := quotedColumnList("clickhouse", cols)
	createSQL := "CREATE TABLE " + qnew + " (" + strings.Join(reordered.items, ", ") + ")" + reordered.suffix
	insertSQL := "INSERT INTO " + qnew + " (" + colList + ") SELECT " + colList + " FROM " + qold

	cleanup := func() {
		tx.Exec("DROP TABLE IF EXISTS " + qnew)
	}
	if err := tx.Exec("DROP TABLE IF EXISTS " + qnew).Error; err != nil {
		return err
	}
	if err := tx.Exec(createSQL).Error; err != nil {
		cleanup()
		return err
	}
	if err := tx.Exec(insertSQL).Error; err != nil {
		cleanup()
		return err
	}
	before, err := countRows(tx, qold)
	if err != nil {
		cleanup()
		return err
	}
	after, err := countRows(tx, qnew)
	if err != nil {
		cleanup()
		return err
	}
	if before != after {
		cleanup()
		return fmt.Errorf("orms: refusing to rebuild %s: copied %d of %d rows", table, after, before)
	}
	rename := "RENAME TABLE " + qold + " TO " + qoldBackup + ", " + qnew + " TO " + qold
	if err := tx.Exec(rename).Error; err != nil {
		cleanup()
		return err
	}
	return tx.Exec("DROP TABLE " + qoldBackup).Error
}

// pgColumn is one row of the pg_attribute metadata used to recreate a column
// with its exact type and constraints.
type pgColumn struct {
	Name        string
	Type        string
	NotNull     bool
	DefaultExpr string
	Identity    string // "a" = ALWAYS, "d" = BY DEFAULT, "" = none
	Generated   string // "s" = STORED
	Collation   string
}

// ddl renders the column definition. It reports whether the column is generated
// (must be excluded from INSERT) and whether it is an identity ALWAYS column
// (requires OVERRIDING SYSTEM VALUE).
func (c pgColumn) ddl() (def string, generated, identityAlways bool) {
	var b strings.Builder
	b.WriteString(quoteIdent("postgres", c.Name))
	b.WriteString(" ")
	b.WriteString(c.Type)
	if c.Collation != "" {
		b.WriteString(" COLLATE ")
		b.WriteString(quoteIdent("postgres", c.Collation))
	}
	switch {
	case c.Generated == "s":
		b.WriteString(" GENERATED ALWAYS AS (")
		b.WriteString(c.DefaultExpr)
		b.WriteString(") STORED")
		return b.String(), true, false
	case c.Identity == "a":
		b.WriteString(" GENERATED ALWAYS AS IDENTITY")
		identityAlways = true
	case c.Identity == "d":
		b.WriteString(" GENERATED BY DEFAULT AS IDENTITY")
	case strings.HasPrefix(c.DefaultExpr, "nextval("):
		// The old sequence is dropped together with the old table, so the
		// default would dangle. Identity has the same semantics and owns its
		// new sequence.
		b.WriteString(" GENERATED BY DEFAULT AS IDENTITY")
	case c.DefaultExpr != "":
		b.WriteString(" DEFAULT ")
		b.WriteString(c.DefaultExpr)
	}
	if c.NotNull {
		b.WriteString(" NOT NULL")
	}
	return b.String(), false, identityAlways
}

type pgConstraint struct {
	Name string
	Type string
	Def  string
}

type pgIndex struct {
	Name string
	Def  string
}

type pgIncomingFK struct {
	Rel  string
	Name string
	Def  string
}

// rebuildPostgres recreates a table with the requested column order and restores
// its primary key, unique/check/foreign-key constraints, indexes, triggers and
// inbound foreign keys. PostgreSQL DDL is transactional, so any failure rolls
// back to the original table.
func rebuildPostgres(tx *gorm.DB, table string, actual, order []string) error {
	toReg := quoteIdent("postgres", table)

	var columns []pgColumn
	if err := tx.Raw(`
		SELECT a.attname AS name,
		       format_type(a.atttypid, a.atttypmod) AS type,
		       a.attnotnull AS not_null,
		       COALESCE(pg_get_expr(d.adbin, d.adrelid), '') AS default_expr,
		       a.attidentity::text AS identity,
		       a.attgenerated::text AS generated,
		       COALESCE(CASE WHEN a.attcollation <> t.typcollation THEN co.collname END, '') AS collation
		FROM pg_attribute a
		JOIN pg_class c ON c.oid = a.attrelid
		JOIN pg_type t ON t.oid = a.atttypid
		LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
		LEFT JOIN pg_collation co ON co.oid = a.attcollation
		WHERE c.oid = to_regclass(?) AND a.attnum > 0 AND NOT a.attisdropped
		ORDER BY a.attnum`, toReg).Scan(&columns).Error; err != nil {
		return fmt.Errorf("orms: read columns of %s: %w", table, err)
	}
	if len(columns) == 0 {
		return fmt.Errorf("orms: table %s has no visible columns", table)
	}
	byName := make(map[string]pgColumn, len(columns))
	for _, c := range columns {
		byName[c.Name] = c
	}

	var pkName, pkDef string
	err := tx.Raw(`SELECT conname, pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = to_regclass(?) AND contype = 'p'`, toReg).
		Row().Scan(&pkName, &pkDef)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("orms: read primary key of %s: %w", table, err)
	}

	var constraints []pgConstraint
	if err := tx.Raw(`
		SELECT conname AS name, contype::text AS type, pg_get_constraintdef(oid) AS def
		FROM pg_constraint
		WHERE conrelid = to_regclass(?) AND contype IN ('u','c','f')
		ORDER BY conname`, toReg).Scan(&constraints).Error; err != nil {
		return fmt.Errorf("orms: read constraints of %s: %w", table, err)
	}

	var indexes []pgIndex
	if err := tx.Raw(`
		SELECT i.relname AS name, pg_get_indexdef(i.oid) AS def
		FROM pg_index x
		JOIN pg_class i ON i.oid = x.indexrelid
		WHERE x.indrelid = to_regclass(?)
		  AND NOT EXISTS (SELECT 1 FROM pg_constraint c WHERE c.conindid = x.indexrelid)
		ORDER BY i.relname`, toReg).Scan(&indexes).Error; err != nil {
		return fmt.Errorf("orms: read indexes of %s: %w", table, err)
	}

	var triggers []pgIndex
	if err := tx.Raw(`
		SELECT tgname AS name, pg_get_triggerdef(oid) AS def
		FROM pg_trigger
		WHERE tgrelid = to_regclass(?) AND NOT tgisinternal
		ORDER BY tgname`, toReg).Scan(&triggers).Error; err != nil {
		return fmt.Errorf("orms: read triggers of %s: %w", table, err)
	}

	var incoming []pgIncomingFK
	if err := tx.Raw(`
		SELECT conrelid::regclass::text AS rel, conname AS name, pg_get_constraintdef(oid) AS def
		FROM pg_constraint
		WHERE contype = 'f' AND confrelid = to_regclass(?)
		ORDER BY conname`, toReg).Scan(&incoming).Error; err != nil {
		return fmt.Errorf("orms: read foreign keys referencing %s: %w", table, err)
	}

	newTable := table + "__sync_new"
	qold := quoteIdent("postgres", table)
	qnew := quoteIdent("postgres", newTable)
	cols := commonColumns(actual, order)

	defs := make([]string, 0, len(cols))
	var insertCols []string
	var identityCols []string
	needOverride := false
	sawInsertable := false
	for _, name := range cols {
		col, ok := byName[name]
		if !ok {
			return fmt.Errorf("orms: column %s missing from %s", name, table)
		}
		def, generated, always := col.ddl()
		defs = append(defs, def)
		if always {
			needOverride = true
		}
		if col.Identity != "" || strings.HasPrefix(col.DefaultExpr, "nextval(") {
			identityCols = append(identityCols, name)
		}
		if generated {
			continue
		}
		insertCols = append(insertCols, name)
		sawInsertable = true
	}
	if !sawInsertable {
		return fmt.Errorf("orms: table %s has no insertable columns", table)
	}

	colList := quotedColumnList("postgres", insertCols)
	insertSQL := "INSERT INTO " + qnew + " (" + colList + ")"
	if needOverride {
		insertSQL += " OVERRIDING SYSTEM VALUE"
	}
	insertSQL += " SELECT " + colList + " FROM " + qold
	createSQL := "CREATE TABLE " + qnew + " (" + strings.Join(defs, ", ") + ")"

	return tx.Transaction(func(tx2 *gorm.DB) error {
		if err := tx2.Exec("DROP TABLE IF EXISTS " + qnew).Error; err != nil {
			return err
		}
		for _, fk := range incoming {
			if err := tx2.Exec("ALTER TABLE " + fk.Rel + " DROP CONSTRAINT " + quoteIdent("postgres", fk.Name)).Error; err != nil {
				return fmt.Errorf("orms: drop inbound foreign key %s: %w", fk.Name, err)
			}
		}
		if err := tx2.Exec(createSQL).Error; err != nil {
			return err
		}
		if err := tx2.Exec(insertSQL).Error; err != nil {
			return err
		}
		before, err := countRows(tx2, qold)
		if err != nil {
			return err
		}
		after, err := countRows(tx2, qnew)
		if err != nil {
			return err
		}
		if before != after {
			return fmt.Errorf("orms: refusing to rebuild %s: copied %d of %d rows", table, after, before)
		}
		if err := tx2.Exec("DROP TABLE " + qold).Error; err != nil {
			return err
		}
		if pkName != "" {
			addPK := "ALTER TABLE " + qnew + " ADD CONSTRAINT " + quoteIdent("postgres", pkName) + " " + pkDef
			if err := tx2.Exec(addPK).Error; err != nil {
				return fmt.Errorf("orms: restore primary key %s: %w", pkName, err)
			}
		}
		if err := tx2.Exec("ALTER TABLE " + qnew + " RENAME TO " + qold).Error; err != nil {
			return err
		}
		for _, idx := range indexes {
			if err := tx2.Exec(idx.Def).Error; err != nil {
				return fmt.Errorf("orms: restore index %s: %w", idx.Name, err)
			}
		}
		for _, con := range constraints {
			add := "ALTER TABLE " + qold + " ADD CONSTRAINT " + quoteIdent("postgres", con.Name) + " " + con.Def
			if err := tx2.Exec(add).Error; err != nil {
				return fmt.Errorf("orms: restore constraint %s: %w", con.Name, err)
			}
		}
		for _, trg := range triggers {
			if err := tx2.Exec(trg.Def).Error; err != nil {
				return fmt.Errorf("orms: restore trigger %s: %w", trg.Name, err)
			}
		}
		for _, fk := range incoming {
			add := "ALTER TABLE " + fk.Rel + " ADD CONSTRAINT " + quoteIdent("postgres", fk.Name) + " " + fk.Def
			if err := tx2.Exec(add).Error; err != nil {
				return fmt.Errorf("orms: restore inbound foreign key %s: %w", fk.Name, err)
			}
		}
		for _, name := range identityCols {
			qcol := quoteIdent("postgres", name)
			seqSQL := fmt.Sprintf(
				"SELECT setval(pg_get_serial_sequence(?, ?), GREATEST(COALESCE(MAX(%s), 1), 1), COALESCE(MAX(%s), 0) > 0) FROM %s",
				qcol, qcol, qold)
			if err := tx2.Exec(seqSQL, qold, name).Error; err != nil {
				return fmt.Errorf("orms: reseed identity column %s: %w", name, err)
			}
		}
		return nil
	})
}

// ssColumn is one sys.columns row.
type ssColumn struct {
	Name        string
	TypeName    string
	MaxLength   int
	Precision   int
	Scale       int
	IsNullable  bool
	IsIdentity  bool
	IsComputed  bool
	DefaultDef  string
	ComputedDef string
	IsPersisted bool
	Collation   string
	Seed        int64
	Increment   int64
}

type ssKey struct {
	Name string
	Kind string
	Cols []string
}

type ssCheck struct {
	Name string
	Def  string
}

type ssIndex struct {
	Name      string
	IsUnique  bool
	TypeDesc  string
	HasFilter bool
	FilterDef string
	KeyCols   []string
	Include   []string
}

type ssFK struct {
	Name         string
	DeleteAct    string
	UpdateAct    string
	RefSchema    string
	RefTable     string
	ParentSchema string
	ParentTable  string
	ParentCol    []string
	RefCol       []string
}

// ssType renders a SQL Server column type, carrying the length/precision that
// sys.columns stores separately from the type name.
func ssType(c ssColumn) string {
	switch c.TypeName {
	case "char", "varchar", "binary", "varbinary":
		if c.MaxLength == -1 {
			return c.TypeName + "(max)"
		}
		return fmt.Sprintf("%s(%d)", c.TypeName, c.MaxLength)
	case "nchar", "nvarchar":
		if c.MaxLength == -1 {
			return c.TypeName + "(max)"
		}
		return fmt.Sprintf("%s(%d)", c.TypeName, c.MaxLength/2)
	case "decimal", "numeric":
		return fmt.Sprintf("%s(%d,%d)", c.TypeName, c.Precision, c.Scale)
	case "datetime2", "datetimeoffset", "time":
		return fmt.Sprintf("%s(%d)", c.TypeName, c.Scale)
	case "float":
		if c.Precision != 53 {
			return fmt.Sprintf("float(%d)", c.Precision)
		}
	}
	return c.TypeName
}

func (c ssColumn) ddl(driver string) string {
	name := quoteIdent(driver, c.Name)
	if c.IsComputed {
		def := name + " AS " + c.ComputedDef
		if c.IsPersisted {
			def += " PERSISTED"
		}
		return def
	}
	var b strings.Builder
	b.WriteString(name)
	b.WriteString(" ")
	b.WriteString(ssType(c))
	if c.Collation != "" {
		b.WriteString(" COLLATE ")
		b.WriteString(c.Collation)
	}
	if c.IsIdentity {
		fmt.Fprintf(&b, " IDENTITY(%d,%d)", c.Seed, c.Increment)
	}
	if c.IsNullable {
		b.WriteString(" NULL")
	} else {
		b.WriteString(" NOT NULL")
	}
	if c.DefaultDef != "" {
		b.WriteString(" DEFAULT ")
		b.WriteString(c.DefaultDef)
	}
	return b.String()
}

// ssAction renders a foreign-key referential action.
func ssAction(desc string) string {
	switch desc {
	case "CASCADE":
		return "CASCADE"
	case "SET_NULL":
		return "SET NULL"
	case "SET_DEFAULT":
		return "SET DEFAULT"
	default: // NO_ACTION
		return ""
	}
}

func ssQualified(schema, table string) string {
	return quoteIdent("sqlserver", schema) + "." + quoteIdent("sqlserver", table)
}

func ssLiteral(s string) string {
	return "N'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// rebuildSQLServer recreates the table in the requested column order and
// restores keys, constraints, indexes, foreign keys and triggers. SQL Server
// DDL is transactional, so a failure rolls back to the original table.
func rebuildSQLServer(tx *gorm.DB, table string, actual, order []string) error {
	driver := "sqlserver"
	obj := quoteIdent(driver, table)

	var columns []ssColumn
	if err := tx.Raw(`
		SELECT c.name AS name,
		       t.name AS type_name,
		       c.max_length AS max_length,
		       c.precision AS precision,
		       c.scale AS scale,
		       c.is_nullable AS is_nullable,
		       c.is_identity AS is_identity,
		       c.is_computed AS is_computed,
		       ISNULL(dc.definition, '') AS default_def,
		       ISNULL(cc.definition, '') AS computed_def,
		       ISNULL(cc.is_persisted, 0) AS is_persisted,
		       ISNULL(c.collation_name, '') AS collation,
		       CAST(ISNULL(ic.seed_value, 1) AS bigint) AS seed,
		       CAST(ISNULL(ic.increment_value, 1) AS bigint) AS increment
		FROM sys.columns c
		JOIN sys.types t ON t.user_type_id = c.user_type_id
		LEFT JOIN sys.default_constraints dc ON dc.object_id = c.default_object_id
		LEFT JOIN sys.computed_columns cc ON cc.object_id = c.object_id AND cc.column_id = c.column_id
		LEFT JOIN sys.identity_columns ic ON ic.object_id = c.object_id AND ic.column_id = c.column_id
		WHERE c.object_id = OBJECT_ID(?)
		ORDER BY c.column_id`, obj).Scan(&columns).Error; err != nil {
		return fmt.Errorf("orms: read columns of %s: %w", table, err)
	}
	if len(columns) == 0 {
		return fmt.Errorf("orms: table %s has no visible columns", table)
	}
	byName := make(map[string]ssColumn, len(columns))
	for _, c := range columns {
		byName[c.Name] = c
	}

	type ssKeyRow struct {
		Name    string
		Kind    string
		Col     string
		DescKey bool
	}
	var keyRows []ssKeyRow
	if err := tx.Raw(`
		SELECT kc.name AS name, kc.type AS kind, c.name AS col, ic.is_descending_key AS desc_key
		FROM sys.key_constraints kc
		JOIN sys.index_columns ic ON ic.object_id = kc.parent_object_id AND ic.index_id = kc.unique_index_id
		JOIN sys.columns c ON c.object_id = ic.object_id AND c.column_id = ic.column_id
		WHERE kc.parent_object_id = OBJECT_ID(?)
		ORDER BY kc.name, ic.key_ordinal`, obj).Scan(&keyRows).Error; err != nil {
		return fmt.Errorf("orms: read keys of %s: %w", table, err)
	}
	var keys []ssKey
	for _, row := range keyRows {
		dir := " ASC"
		if row.DescKey {
			dir = " DESC"
		}
		col := quoteIdent(driver, row.Col) + dir
		if n := len(keys); n > 0 && keys[n-1].Name == row.Name {
			keys[n-1].Cols = append(keys[n-1].Cols, col)
			continue
		}
		keys = append(keys, ssKey{Name: row.Name, Kind: row.Kind, Cols: []string{col}})
	}

	var checks []ssCheck
	if err := tx.Raw(`
		SELECT cc.name AS name, cc.definition AS def
		FROM sys.check_constraints cc
		WHERE cc.parent_object_id = OBJECT_ID(?)
		ORDER BY cc.name`, obj).Scan(&checks).Error; err != nil {
		return fmt.Errorf("orms: read check constraints of %s: %w", table, err)
	}

	type ssIndexRow struct {
		ID        int
		Name      string
		IsUnique  bool
		TypeDesc  string
		HasFilter bool
		FilterDef string
	}
	var indexRows []ssIndexRow
	if err := tx.Raw(`
		SELECT i.index_id AS id, i.name AS name, i.is_unique AS is_unique,
		       i.type_desc AS type_desc, i.has_filter AS has_filter,
		       ISNULL(i.filter_definition, '') AS filter_def
		FROM sys.indexes i
		WHERE i.object_id = OBJECT_ID(?) AND i.is_primary_key = 0 AND i.is_unique_constraint = 0
		  AND i.is_hypothetical = 0 AND i.name IS NOT NULL
		ORDER BY i.name`, obj).Scan(&indexRows).Error; err != nil {
		return fmt.Errorf("orms: read indexes of %s: %w", table, err)
	}
	indexes := make([]ssIndex, 0, len(indexRows))
	for _, row := range indexRows {
		switch row.TypeDesc {
		case "NONCLUSTERED":
		case "CLUSTERED":
		default:
			// Column-store, XML, spatial or hash indexes cannot be rebuilt from
			// sys.index_columns; fail before touching the table.
			return fmt.Errorf("orms: cannot rebuild %s index %s on %s", row.TypeDesc, row.Name, table)
		}
		idx := ssIndex{
			Name:      row.Name,
			IsUnique:  row.IsUnique,
			TypeDesc:  row.TypeDesc,
			HasFilter: row.HasFilter,
			FilterDef: row.FilterDef,
		}
		type ssIndexCol struct {
			Name     string
			DescKey  bool
			Included bool
		}
		var cols []ssIndexCol
		if err := tx.Raw(`
			SELECT c.name AS name, ic.is_descending_key AS desc_key, ic.is_included_column AS included
			FROM sys.index_columns ic
			JOIN sys.columns c ON c.object_id = ic.object_id AND c.column_id = ic.column_id
			WHERE ic.object_id = OBJECT_ID(?) AND ic.index_id = ?
			ORDER BY ic.is_included_column, ic.key_ordinal, ic.index_column_id`, obj, row.ID).Scan(&cols).Error; err != nil {
			return fmt.Errorf("orms: read index %s columns: %w", row.Name, err)
		}
		for _, col := range cols {
			if col.Included {
				idx.Include = append(idx.Include, quoteIdent(driver, col.Name))
				continue
			}
			dir := " ASC"
			if col.DescKey {
				dir = " DESC"
			}
			idx.KeyCols = append(idx.KeyCols, quoteIdent(driver, col.Name)+dir)
		}
		indexes = append(indexes, idx)
	}

	type ssFKRow struct {
		Name         string
		DelAction    string
		UpdAction    string
		RefSchema    string
		RefTable     string
		ParentRel    string
		ParentCol    string
		RefCol       string
		ParentSchema string
		ParentTable  string
	}
	readFKs := func(where string) ([]ssFK, error) {
		var rows []ssFKRow
		if err := tx.Raw(`
			SELECT fk.name AS name,
			       fk.delete_referential_action_desc AS del_action,
			       fk.update_referential_action_desc AS upd_action,
			       OBJECT_SCHEMA_NAME(fk.referenced_object_id) AS ref_schema,
			       OBJECT_NAME(fk.referenced_object_id) AS ref_table,
			       OBJECT_SCHEMA_NAME(fk.parent_object_id) AS parent_schema,
			       OBJECT_NAME(fk.parent_object_id) AS parent_table,
			       pc.name AS parent_col, rc.name AS ref_col
			FROM sys.foreign_keys fk
			JOIN sys.foreign_key_columns fkc ON fkc.constraint_object_id = fk.object_id
			JOIN sys.columns pc ON pc.object_id = fkc.parent_object_id AND pc.column_id = fkc.parent_column_id
			JOIN sys.columns rc ON rc.object_id = fkc.referenced_object_id AND rc.column_id = fkc.referenced_column_id
			WHERE `+where+`
			ORDER BY fk.name, fkc.constraint_column_id`, obj).Scan(&rows).Error; err != nil {
			return nil, err
		}
		var out []ssFK
		for _, row := range rows {
			if n := len(out); n > 0 && out[n-1].Name == row.Name {
				out[n-1].ParentCol = append(out[n-1].ParentCol, row.ParentCol)
				out[n-1].RefCol = append(out[n-1].RefCol, row.RefCol)
				continue
			}
			out = append(out, ssFK{
				Name:         row.Name,
				DeleteAct:    row.DelAction,
				UpdateAct:    row.UpdAction,
				RefSchema:    row.RefSchema,
				RefTable:     row.RefTable,
				ParentSchema: row.ParentSchema,
				ParentTable:  row.ParentTable,
				ParentCol:    []string{row.ParentCol},
				RefCol:       []string{row.RefCol},
			})
		}
		return out, nil
	}
	outgoing, err := readFKs("fk.parent_object_id = OBJECT_ID(?)")
	if err != nil {
		return fmt.Errorf("orms: read foreign keys of %s: %w", table, err)
	}
	incoming, err := readFKs("fk.referenced_object_id = OBJECT_ID(?)")
	if err != nil {
		return fmt.Errorf("orms: read foreign keys referencing %s: %w", table, err)
	}

	type ssTrigger struct {
		Name string
		Def  string
	}
	var triggers []ssTrigger
	if err := tx.Raw(`
		SELECT tr.name AS name, OBJECT_DEFINITION(tr.object_id) AS def
		FROM sys.triggers tr
		WHERE tr.parent_id = OBJECT_ID(?)
		ORDER BY tr.name`, obj).Scan(&triggers).Error; err != nil {
		return fmt.Errorf("orms: read triggers of %s: %w", table, err)
	}

	newTable := table + "__sync_new"
	qold := quoteIdent(driver, table)
	qnew := quoteIdent(driver, newTable)
	cols := commonColumns(actual, order)

	defs := make([]string, 0, len(cols))
	var insertCols []string
	hasIdentity := false
	for _, name := range cols {
		col, ok := byName[name]
		if !ok {
			return fmt.Errorf("orms: column %s missing from %s", name, table)
		}
		defs = append(defs, col.ddl(driver))
		if col.IsIdentity {
			hasIdentity = true
		}
		if !col.IsComputed {
			insertCols = append(insertCols, name)
		}
	}
	if len(insertCols) == 0 {
		return fmt.Errorf("orms: table %s has no insertable columns", table)
	}
	colList := quotedColumnList(driver, insertCols)
	createSQL := "CREATE TABLE " + qnew + " (" + strings.Join(defs, ", ") + ")"
	insertSQL := "INSERT INTO " + qnew + " (" + colList + ") SELECT " + colList + " FROM " + qold

	return tx.Transaction(func(tx2 *gorm.DB) error {
		if err := tx2.Exec("DROP TABLE IF EXISTS " + qnew).Error; err != nil {
			return err
		}
		if err := tx2.Exec(createSQL).Error; err != nil {
			return err
		}
		if hasIdentity {
			if err := tx2.Exec("SET IDENTITY_INSERT " + qnew + " ON").Error; err != nil {
				return err
			}
		}
		if err := tx2.Exec(insertSQL).Error; err != nil {
			return err
		}
		if hasIdentity {
			if err := tx2.Exec("SET IDENTITY_INSERT " + qnew + " OFF").Error; err != nil {
				return err
			}
		}
		before, err := countRows(tx2, qold)
		if err != nil {
			return err
		}
		after, err := countRows(tx2, qnew)
		if err != nil {
			return err
		}
		if before != after {
			return fmt.Errorf("orms: refusing to rebuild %s: copied %d of %d rows", table, after, before)
		}
		for _, fk := range incoming {
			drop := "ALTER TABLE " + ssQualified(fk.ParentSchema, fk.ParentTable) + " DROP CONSTRAINT " + quoteIdent(driver, fk.Name)
			if err := tx2.Exec(drop).Error; err != nil {
				return fmt.Errorf("orms: drop inbound foreign key %s: %w", fk.Name, err)
			}
		}
		if err := tx2.Exec("DROP TABLE " + qold).Error; err != nil {
			return err
		}
		// sp_rename only strips brackets from the *old* name: bracketing the new
		// name creates a table literally named "[table]".
		rename := "EXEC sp_rename " + ssLiteral(quoteIdent(driver, newTable)) + ", " + ssLiteral(table) + ", 'OBJECT'"
		if err := tx2.Exec(rename).Error; err != nil {
			return err
		}
		for _, key := range keys {
			kind := "UNIQUE"
			if key.Kind == "PK" {
				kind = "PRIMARY KEY"
			}
			add := "ALTER TABLE " + qold + " ADD CONSTRAINT " + quoteIdent(driver, key.Name) + " " + kind + " (" + strings.Join(key.Cols, ", ") + ")"
			if err := tx2.Exec(add).Error; err != nil {
				return fmt.Errorf("orms: restore constraint %s: %w", key.Name, err)
			}
		}
		for _, check := range checks {
			add := "ALTER TABLE " + qold + " ADD CONSTRAINT " + quoteIdent(driver, check.Name) + " CHECK " + check.Def
			if err := tx2.Exec(add).Error; err != nil {
				return fmt.Errorf("orms: restore check constraint %s: %w", check.Name, err)
			}
		}
		for _, idx := range indexes {
			var b strings.Builder
			b.WriteString("CREATE ")
			if idx.IsUnique {
				b.WriteString("UNIQUE ")
			}
			b.WriteString("INDEX ")
			b.WriteString(quoteIdent(driver, idx.Name))
			b.WriteString(" ON ")
			b.WriteString(qold)
			b.WriteString(" (")
			b.WriteString(strings.Join(idx.KeyCols, ", "))
			b.WriteString(")")
			if len(idx.Include) > 0 {
				b.WriteString(" INCLUDE (")
				b.WriteString(strings.Join(idx.Include, ", "))
				b.WriteString(")")
			}
			if idx.HasFilter && idx.FilterDef != "" {
				b.WriteString(" WHERE ")
				b.WriteString(idx.FilterDef)
			}
			if err := tx2.Exec(b.String()).Error; err != nil {
				return fmt.Errorf("orms: restore index %s: %w", idx.Name, err)
			}
		}
		addFK := func(parent, ref string, fk ssFK) error {
			var b strings.Builder
			b.WriteString("ALTER TABLE ")
			b.WriteString(parent)
			b.WriteString(" ADD CONSTRAINT ")
			b.WriteString(quoteIdent(driver, fk.Name))
			b.WriteString(" FOREIGN KEY (")
			for i, col := range fk.ParentCol {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(quoteIdent(driver, col))
			}
			b.WriteString(") REFERENCES ")
			b.WriteString(ref)
			b.WriteString(" (")
			for i, col := range fk.RefCol {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(quoteIdent(driver, col))
			}
			b.WriteString(")")
			if action := ssAction(fk.DeleteAct); action != "" {
				b.WriteString(" ON DELETE ")
				b.WriteString(action)
			}
			if action := ssAction(fk.UpdateAct); action != "" {
				b.WriteString(" ON UPDATE ")
				b.WriteString(action)
			}
			if err := tx2.Exec(b.String()).Error; err != nil {
				return fmt.Errorf("orms: restore foreign key %s: %w", fk.Name, err)
			}
			return nil
		}
		for _, fk := range outgoing {
			if err := addFK(qold, ssQualified(fk.RefSchema, fk.RefTable), fk); err != nil {
				return err
			}
		}
		for _, fk := range incoming {
			if err := addFK(ssQualified(fk.ParentSchema, fk.ParentTable), qold, fk); err != nil {
				return err
			}
		}
		for _, trg := range triggers {
			stmt := "EXEC sp_executesql " + ssLiteral(trg.Def)
			if err := tx2.Exec(stmt).Error; err != nil {
				return fmt.Errorf("orms: restore trigger %s: %w", trg.Name, err)
			}
		}
		if hasIdentity {
			if err := tx2.Exec("DBCC CHECKIDENT (" + ssLiteral(quoteIdent(driver, table)) + ", RESEED)").Error; err != nil {
				return fmt.Errorf("orms: reseed identity of %s: %w", table, err)
			}
		}
		return nil
	})
}
