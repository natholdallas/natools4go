package orms

import "fmt"

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
