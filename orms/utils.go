package orms

import (
	"fmt"

	"gorm.io/gorm"
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
