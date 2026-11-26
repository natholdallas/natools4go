//go:build integration

package orms

import (
	"fmt"
	"os"
	"slices"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// integV1 is the initially migrated shape: id, name, tags.
type integV1 struct {
	ID   int `gorm:"primaryKey"`
	Name string
	Tags List[string]
}

func (integV1) TableName() string { return "sync_integration" }

// integV2 declares the columns in a different order, so SyncTablesSoft has to
// reorder the physical table.
type integV2 struct {
	ID   int `gorm:"primaryKey"`
	Tags List[string]
	Name string
}

func (integV2) TableName() string { return "sync_integration" }

const integDBName = "natools_test"

func TestSyncTablesEveryDriver(t *testing.T) {
	cases := []struct {
		name   string
		driver string
		env    string
	}{
		{"mysql", "mysql", "ORMS_MYSQL_DSN"},
		{"postgres", "postgres", "ORMS_POSTGRES_DSN"},
		{"sqlserver", "sqlserver", "ORMS_SQLSERVER_DSN"},
		{"clickhouse", "clickhouse", "ORMS_CLICKHOUSE_DSN"},
	}
	ran := 0
	for _, tc := range cases {
		dsn := os.Getenv(tc.env)
		if dsn == "" {
			t.Logf("%s not set, skipping %s", tc.env, tc.name)
			continue
		}
		ran++
		t.Run(tc.name, func(t *testing.T) {
			runSyncScenario(t, tc.driver, dsn)
		})
	}
	if ran == 0 {
		t.Skip("no ORMS_*_DSN variables set")
	}
}

func runSyncScenario(t *testing.T, driver, dsn string) {
	t.Helper()

	// AutoCreate must be idempotent.
	if err := AutoCreateE(integDBName, driver, dsn); err != nil {
		t.Fatalf("AutoCreateE (first): %v", err)
	}
	if err := AutoCreateE(integDBName, driver, dsn); err != nil {
		t.Fatalf("AutoCreateE (second): %v", err)
	}

	d, err := Dialector(driver, dsn, integDBName, "")
	if err != nil {
		t.Fatalf("Dialector: %v", err)
	}
	db, err := New(d, &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("DB: %v", err)
	}
	defer sqlDB.Close()

	// Clean slate.
	for _, table := range []string{"sync_child", "sync_integration"} {
		if err := db.Migrator().DropTable(table); err != nil {
			t.Fatalf("drop %s: %v", table, err)
		}
	}

	if err := AutoMigrate(db, &integV1{}); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}

	// Seed three rows. ClickHouse has no auto-increment, so ids are explicit;
	// the List field still goes through driver.Valuer in both cases.
	if driver == "clickhouse" {
		for i := 1; i <= 3; i++ {
			row := integV1{ID: i, Name: fmt.Sprintf("n%d", i), Tags: List[string]{fmt.Sprintf("t%d", i)}}
			if err := db.Create(&row).Error; err != nil {
				t.Fatalf("seed clickhouse: %v", err)
			}
		}
	} else {
		for i := 1; i <= 3; i++ {
			row := integV1{Name: fmt.Sprintf("n%d", i), Tags: List[string]{fmt.Sprintf("t%d", i)}}
			if err := db.Create(&row).Error; err != nil {
				t.Fatalf("seed: %v", err)
			}
		}
	}

	// A database-only column that soft mode must keep, sorted last.
	addColumn := map[string]string{
		"mysql":      "ALTER TABLE sync_integration ADD COLUMN zzz TEXT",
		"mariadb":    "ALTER TABLE sync_integration ADD COLUMN zzz TEXT",
		"postgres":   "ALTER TABLE sync_integration ADD COLUMN zzz text",
		"sqlserver":  "ALTER TABLE sync_integration ADD zzz nvarchar(max)",
		"clickhouse": "ALTER TABLE sync_integration ADD COLUMN zzz String",
	}[driver]
	if addColumn == "" {
		t.Fatalf("no add-column statement for driver %s", driver)
	}
	if err := db.Exec(addColumn).Error; err != nil {
		t.Fatalf("add extra column: %v", err)
	}

	// Index and inbound foreign key are supported by every driver except
	// ClickHouse (no enforced foreign keys, different index model).
	hasFK := driver != "clickhouse"
	if driver == "mysql" || driver == "mariadb" {
		// GORM maps string to TEXT, which needs a prefix length in an index.
		if err := db.Exec("ALTER TABLE sync_integration MODIFY COLUMN name varchar(100) NULL").Error; err != nil {
			t.Fatalf("narrow name column: %v", err)
		}
	}
	if driver == "sqlserver" {
		// GORM maps string to nvarchar(max), which cannot be an index key.
		if err := db.Exec("ALTER TABLE sync_integration ALTER COLUMN name nvarchar(100) NULL").Error; err != nil {
			t.Fatalf("narrow name column: %v", err)
		}
	}
	if driver == "mysql" || driver == "mariadb" || driver == "postgres" || driver == "sqlserver" {
		if err := db.Exec("CREATE INDEX idx_sync_name ON sync_integration (name)").Error; err != nil {
			t.Fatalf("create index: %v", err)
		}
	}
	if hasFK {
		var childDDL string
		switch driver {
		case "mysql", "mariadb":
			childDDL = "CREATE TABLE sync_child (id INT PRIMARY KEY AUTO_INCREMENT, parent_id BIGINT, CONSTRAINT fk_sync_child FOREIGN KEY (parent_id) REFERENCES sync_integration(id))"
		case "postgres":
			childDDL = "CREATE TABLE sync_child (id serial PRIMARY KEY, parent_id integer, CONSTRAINT fk_sync_child FOREIGN KEY (parent_id) REFERENCES sync_integration(id))"
		case "sqlserver":
			childDDL = "CREATE TABLE sync_child (id int IDENTITY(1,1) PRIMARY KEY, parent_id bigint, CONSTRAINT fk_sync_child FOREIGN KEY (parent_id) REFERENCES sync_integration(id))"
		}
		if err := db.Exec(childDDL).Error; err != nil {
			t.Fatalf("create child table: %v", err)
		}
		if err := db.Exec("INSERT INTO sync_child (parent_id) VALUES (1)").Error; err != nil {
			t.Fatalf("seed child: %v", err)
		}
	}

	// The actual reorder: struct order is now id, tags, name.
	if err := SyncTablesSoft(db, &integV2{}); err != nil {
		t.Fatalf("SyncTablesSoft: %v", err)
	}

	want := []string{"id", "tags", "name", "zzz"}
	if got := integColumnOrder(t, db, driver); !slices.Equal(got, want) {
		t.Errorf("column order = %v, want %v", got, want)
	}

	var count int64
	if err := db.Raw("SELECT count(*) FROM sync_integration").Scan(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 3 {
		t.Errorf("rows after sync = %d, want 3", count)
	}

	var name, tags string
	if err := db.Raw("SELECT name, tags FROM sync_integration WHERE id = 2").Row().Scan(&name, &tags); err != nil {
		t.Fatalf("read row: %v", err)
	}
	if name != "n2" {
		t.Errorf("name = %q, want n2", name)
	}

	if driver != "clickhouse" {
		// Identity must continue after the highest copied id.
		row := integV2{Name: "n4", Tags: List[string]{"t4"}}
		if err := db.Create(&row).Error; err != nil {
			t.Fatalf("insert after sync: %v", err)
		}
		if row.ID <= 3 {
			t.Errorf("new id = %d, want > 3", row.ID)
		}

		// The index must still exist and be usable.
		if driver == "mysql" || driver == "mariadb" || driver == "postgres" || driver == "sqlserver" {
			var n int64
			idxQuery := map[string]string{
				"mysql":     "SELECT count(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'sync_integration' AND index_name = 'idx_sync_name'",
				"mariadb":   "SELECT count(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'sync_integration' AND index_name = 'idx_sync_name'",
				"postgres":  "SELECT count(*) FROM pg_indexes WHERE tablename = 'sync_integration' AND indexname = 'idx_sync_name'",
				"sqlserver": "SELECT count(*) FROM sys.indexes WHERE object_id = OBJECT_ID('sync_integration') AND name = 'idx_sync_name'",
			}[driver]
			if err := db.Raw(idxQuery).Scan(&n).Error; err != nil {
				t.Fatalf("check index: %v", err)
			}
			if n != 1 {
				t.Errorf("index idx_sync_name missing after rebuild")
			}
		}
	}

	if hasFK {
		// The child row must survive and the FK must still be enforced.
		var childCount int64
		if err := db.Raw("SELECT count(*) FROM sync_child").Scan(&childCount).Error; err != nil {
			t.Fatalf("count child: %v", err)
		}
		if childCount != 1 {
			t.Errorf("child rows = %d, want 1", childCount)
		}
		if err := db.Exec("INSERT INTO sync_child (parent_id) VALUES (999)").Error; err == nil {
			t.Error("foreign key was not restored: orphan insert succeeded")
		}
	}

	// A second sync must not run any DDL.
	rec := &sqlRecorder{Interface: logger.Discard}
	db.Logger = rec
	if err := SyncTablesSoft(db, &integV2{}); err != nil {
		t.Fatalf("second SyncTablesSoft: %v", err)
	}
	if n := rec.ddlCount(); n != 0 {
		t.Errorf("second sync emitted %d DDL statements: %v", n, rec.sqls)
	}
}

func integColumnOrder(t *testing.T, db *gorm.DB, driver string) []string {
	t.Helper()
	var names []string
	var query string
	switch driver {
	case "mysql", "mariadb":
		query = "SELECT column_name FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'sync_integration' ORDER BY ordinal_position"
	case "postgres":
		query = "SELECT attname FROM pg_attribute WHERE attrelid = 'sync_integration'::regclass AND attnum > 0 AND NOT attisdropped ORDER BY attnum"
	case "sqlserver":
		query = "SELECT name FROM sys.columns WHERE object_id = OBJECT_ID('sync_integration') ORDER BY column_id"
	case "clickhouse":
		query = "SELECT name FROM system.columns WHERE database = currentDatabase() AND table = 'sync_integration' ORDER BY position"
	default:
		t.Fatalf("no column query for driver %s", driver)
	}
	if err := db.Raw(query).Scan(&names).Error; err != nil {
		t.Fatalf("read column order: %v", err)
	}
	return names
}

// TestResetEveryDriver proves Reset (drop + create database) is valid SQL on
// every server driver and leaves a usable database behind.
func TestResetEveryDriver(t *testing.T) {
	cases := []struct {
		name   string
		driver string
		env    string
	}{
		{"mysql", "mysql", "ORMS_MYSQL_DSN"},
		{"postgres", "postgres", "ORMS_POSTGRES_DSN"},
		{"sqlserver", "sqlserver", "ORMS_SQLSERVER_DSN"},
		{"clickhouse", "clickhouse", "ORMS_CLICKHOUSE_DSN"},
	}
	ran := 0
	for _, tc := range cases {
		dsn := os.Getenv(tc.env)
		if dsn == "" {
			continue
		}
		ran++
		t.Run(tc.name, func(t *testing.T) {
			const dbName = "natools_reset_test"
			if err := Reset(dbName, tc.driver, dsn); err != nil {
				t.Fatalf("Reset (first): %v", err)
			}
			d, err := Dialector(tc.driver, dsn, dbName, "")
			if err != nil {
				t.Fatalf("Dialector: %v", err)
			}
			db, err := New(d, &gorm.Config{Logger: logger.Discard})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			if err := AutoMigrate(db, &integV1{}); err != nil {
				t.Fatalf("AutoMigrate after reset: %v", err)
			}
			sqlDB, err := db.DB()
			if err != nil {
				t.Fatalf("DB: %v", err)
			}
			sqlDB.Close()
			// Dropping a database that already has tables is the real path; a
			// second Reset must still succeed.
			if err := Reset(dbName, tc.driver, dsn); err != nil {
				t.Fatalf("Reset (second): %v", err)
			}
		})
	}
	if ran == 0 {
		t.Skip("no ORMS_*_DSN variables set")
	}
}
