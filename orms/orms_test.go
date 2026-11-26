package orms

import (
	"context"
	"io"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

type user struct {
	Model[int]
	Name  string
	Age   int
	Tags  List[string]
	Meta  Dict[int]
	Email string
}

func TestValidateDBName(t *testing.T) {
	for _, ok := range []string{"app", "my_db", "MyDB1", "a_b_c"} {
		if err := validateDBName(ok); err != nil {
			t.Errorf("validateDBName(%q) = %v, want nil", ok, err)
		}
	}
	for _, bad := range []string{"", "a b", "app;DROP TABLE x", "a-b", "1abc", "app\" OR 1=1", "db."} {
		if err := validateDBName(bad); err == nil {
			t.Errorf("validateDBName(%q) = nil, want error", bad)
		}
	}
}

func TestMigrateAndCRUD(t *testing.T) {
	db := newTestDB(t)
	if err := db.AutoMigrate(&user{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	u := user{Name: "alice", Age: 30, Tags: List[string]{"a", "b"}, Meta: Dict[int]{"score": 9}}
	if err := Create(db, &u); err != nil {
		t.Fatalf("create: %v", err)
	}
	if u.ID == 0 {
		t.Error("expected auto-generated ID")
	}

	got, err := First[user](db, u.ID)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	if got.Name != "alice" || got.Age != 30 {
		t.Errorf("got %+v, want alice/30", got)
	}
	if len(got.Tags) != 2 || got.Meta["score"] != 9 {
		t.Errorf("json cols not round-tripped: %+v", got)
	}
}

func TestQueryBuilder(t *testing.T) {
	db := newTestDB(t)
	_ = db.AutoMigrate(&user{})
	seedUsers(t, db, []string{"a", "b", "c"})

	q := QE[user](db)
	list, err := q.Where("age > ?", 1).Order("age desc").Find()
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("list len = %d, want 2", len(list))
	}
	if list[0].Age < list[1].Age {
		t.Error("expected descending order")
	}

	u, err := QE[user](db).Where("name = ?", "b").First()
	if err != nil || u.Name != "b" {
		t.Errorf("first = %+v, %v; want b", u, err)
	}

	if got := QE[user](db).IFind(); len(got) != 3 {
		t.Errorf("IFind len = %d, want 3", len(got))
	}
}

func TestCountExists(t *testing.T) {
	db := newTestDB(t)
	_ = db.AutoMigrate(&user{})
	seedUsers(t, db, []string{"a", "b", "c"})

	if got := Count[user](db); got != 3 {
		t.Errorf("Count = %d, want 3", got)
	}
	if !Exists[user](db) {
		t.Error("Exists should be true")
	}
	if got := QE[user](db).Count(); got != 3 {
		t.Errorf("Query.Count = %d, want 3", got)
	}
}

func TestPaginate(t *testing.T) {
	db := newTestDB(t)
	_ = db.AutoMigrate(&user{})
	seedUsers(t, db, []string{"a", "b", "c", "d", "e"})

	page, _ := QE[user](db).Paginate(Pagination{Page: 2, Size: 2})
	if page.Total != 5 {
		t.Errorf("Total = %d, want 5", page.Total)
	}
	if page.Page != 3 {
		t.Errorf("Page = %d, want 3", page.Page)
	}
	if len(page.Content) != 2 {
		t.Errorf("Content len = %d, want 2", len(page.Content))
	}
}

func TestPaginateNormalizedSize(t *testing.T) {
	db := newTestDB(t)
	_ = db.AutoMigrate(&user{})
	seedUsers(t, db, []string{"a", "b", "c", "d", "e"})

	// Size=0 is normalized to 20 by the scope, so the page count must be
	// computed against 20 (regression: it used the raw size 0).
	page, _ := QE[user](db).Paginate(Pagination{Page: 1, Size: 0})
	if page.Page != 1 {
		t.Errorf("Page with Size=0 = %d, want 1", page.Page)
	}

	// Oversized Size is clamped to 100; the page count must match.
	page, _ = QE[user](db).Paginate(Pagination{Page: 1, Size: 200})
	if page.Page != 1 {
		t.Errorf("Page with Size=200 = %d, want 1", page.Page)
	}
	if len(page.Content) != 5 {
		t.Errorf("Content len = %d, want 5", len(page.Content))
	}
}

type codeUser struct {
	Code string `gorm:"column:code;primaryKey"`
	Name string
}

func TestUpdateByIDCustomPrimaryKey(t *testing.T) {
	db := newTestDB(t)
	if err := db.AutoMigrate(&codeUser{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := Create(db, &codeUser{Code: "u-1", Name: "alice"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	// Regression: UpdateByID used to hardcode "id" as the primary key,
	// silently no-op on models whose PK has a different column name.
	if err := UpdateByID[codeUser](db, "u-1", "name", "bob"); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, err := First[codeUser](db, "code = ?", "u-1")
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if got.Name != "bob" {
		t.Errorf("name = %q, want bob", got.Name)
	}
}

func TestPaginationScopeSideEffectFree(t *testing.T) {
	db := newTestDB(t)
	_ = db.AutoMigrate(&user{})

	p := Pagination{Page: 0, Size: 0}
	p.Scope(db)
	if p.Page != 0 || p.Size != 0 {
		t.Errorf("Scope mutated receiver: %+v", p)
	}
}

func TestSorters(t *testing.T) {
	db := newTestDB(t)
	_ = db.AutoMigrate(&user{})
	seedUsers(t, db, []string{"a", "b", "c"})

	s := Sorters{Columns: []Sorter{{Column: "age", Desc: true}}}
	list, err := QE[user](db).Scopes(s.Scope).Find()
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if list[0].Age < list[len(list)-1].Age {
		t.Error("expected descending sort")
	}
}

func TestSorterColumnValidation(t *testing.T) {
	for _, ok := range []string{"id", "created_at", "Name", "a1"} {
		if !validColumn(ok) {
			t.Errorf("validColumn(%q) = false, want true", ok)
		}
	}
	for _, bad := range []string{"", "age;DROP", "a b", "1x", "col.name"} {
		if validColumn(bad) {
			t.Errorf("validColumn(%q) = true, want false", bad)
		}
	}

	db := newTestDB(t)
	_ = db.AutoMigrate(&user{})
	seedUsers(t, db, []string{"a", "b", "c"})

	// An invalid sort column must be ignored (no error, no injection).
	s := Sorter{Column: "age DESC; DROP TABLE users; --"}
	got, err := QE[user](db).Scopes(s.Scope).Find()
	if err != nil {
		t.Fatalf("find with invalid sorter: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("got %d users, want 3", len(got))
	}
}

func TestUpdateDelete(t *testing.T) {
	db := newTestDB(t)
	_ = db.AutoMigrate(&user{})
	seedUsers(t, db, []string{"a", "b"})

	if err := UpdateByID[user](db, 1, "age", 99); err != nil {
		t.Fatalf("update: %v", err)
	}
	if got := IFirst[user](db, 1).Age; got != 99 {
		t.Errorf("age = %d, want 99", got)
	}

	if err := Delete[user](db, 1); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if got := Count[user](db); got != 1 {
		t.Errorf("Count after delete = %d, want 1", got)
	}
}

func TestIFirstZeroWhenMissing(t *testing.T) {
	db := newTestDB(t)
	_ = db.AutoMigrate(&user{})
	if got := IFirst[user](db, 999); got.ID != 0 {
		t.Errorf("IFirst missing = %+v, want zero", got)
	}
}

func TestLogPreset(t *testing.T) {
	l := LogPreset(io.Discard, logger.Warn)
	if l == nil {
		t.Fatal("LogPreset returned nil")
	}
}

func TestDSN(t *testing.T) {
	cases := []struct {
		driver string
		want   string
	}{
		{"mysql", "u:p@tcp(h:1)/?parseTime=true&loc=Local"},
		{"postgres", "host=h user=u password=p port=1"},
		{"sqlite", "db"},
		{"sqlserver", "sqlserver://u:p@h:1"},
		{"clickhouse", "tcp://h:1?username=u&password=p"},
	}
	for _, c := range cases {
		got, err := DSN(c.driver, "db", "u", "p", "h", "1")
		if err != nil {
			t.Errorf("DSN(%s) error: %v", c.driver, err)
			continue
		}
		if got != c.want {
			t.Errorf("DSN(%s) = %q, want %q", c.driver, got, c.want)
		}
	}
	if _, err := DSN("unknown", "db", "u", "p", "h", "1"); err == nil {
		t.Error("expected error for unsupported driver")
	}
}

func TestJSONScanValue(t *testing.T) {
	var l List[int]
	if err := JSONScan(&l, []byte("[1,2,3]")); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(l) != 3 || l[0] != 1 {
		t.Errorf("List = %v, want [1 2 3]", l)
	}

	v, err := JSONValue(l)
	if err != nil || string(v.([]byte)) != "[1,2,3]" {
		t.Errorf("value = %v, %v", v, err)
	}

	// nil clears target
	l = List[int]{9}
	if err := JSONScan(&l, nil); err != nil {
		t.Fatalf("scan nil: %v", err)
	}
	if l != nil {
		t.Errorf("List after nil scan = %v, want nil", l)
	}
}

func seedUsers(t *testing.T, db *gorm.DB, names []string) {
	t.Helper()
	for i, n := range names {
		u := user{Name: n, Age: i + 1, Email: n + "@x.com"}
		if err := db.Create(&u).Error; err != nil {
			t.Fatalf("seed %s: %v", n, err)
		}
		time.Sleep(time.Millisecond) // ensure distinct created_at for deterministic ordering
	}
}

type syncUser struct {
	ID   int `gorm:"primaryKey"`
	Name string
	Age  int
}

func (syncUser) TableName() string { return "sync_users" }

func tableColumns(t *testing.T, db *gorm.DB, table string) []string {
	t.Helper()
	var cols []struct{ Name string }
	if err := db.Raw("PRAGMA table_info(" + table + ")").Scan(&cols).Error; err != nil {
		t.Fatalf("read columns of %s: %v", table, err)
	}
	names := make([]string, len(cols))
	for i, c := range cols {
		names[i] = c.Name
	}
	return names
}

func TestSyncTablesHard(t *testing.T) {
	db := newTestDB(t)
	if err := db.AutoMigrate(&syncUser{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Exec("ALTER TABLE sync_users ADD COLUMN legacy TEXT").Error; err != nil {
		t.Fatalf("add legacy column: %v", err)
	}
	if err := db.Exec("INSERT INTO sync_users (name, age, legacy) VALUES (?, ?, ?)", "alice", 30, "keep").Error; err != nil {
		t.Fatalf("insert: %v", err)
	}

	if err := SyncTablesHard(db, &syncUser{}); err != nil {
		t.Fatalf("SyncTablesHard: %v", err)
	}

	cols := tableColumns(t, db, "sync_users")
	for _, c := range cols {
		if c == "legacy" {
			t.Errorf("hard sync kept dropped column legacy: %v", cols)
		}
	}

	var name string
	var age int
	if err := db.Raw("SELECT name, age FROM sync_users WHERE id = 1").Row().Scan(&name, &age); err != nil {
		t.Fatalf("read row: %v", err)
	}
	if name != "alice" || age != 30 {
		t.Errorf("data lost after hard sync: name=%q age=%d", name, age)
	}
}

func TestSyncTablesSoft(t *testing.T) {
	db := newTestDB(t)
	if err := db.AutoMigrate(&syncUser{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Exec("ALTER TABLE sync_users ADD COLUMN zzz TEXT").Error; err != nil {
		t.Fatalf("add zzz column: %v", err)
	}
	if err := db.Exec("ALTER TABLE sync_users ADD COLUMN aaa TEXT").Error; err != nil {
		t.Fatalf("add aaa column: %v", err)
	}
	if err := db.Exec("INSERT INTO sync_users (name, age, zzz, aaa) VALUES (?, ?, ?, ?)", "bob", 25, "z-data", "a-data").Error; err != nil {
		t.Fatalf("insert: %v", err)
	}

	if err := SyncTablesSoft(db, &syncUser{}); err != nil {
		t.Fatalf("SyncTablesSoft: %v", err)
	}

	cols := tableColumns(t, db, "sync_users")
	want := []string{"id", "name", "age", "aaa", "zzz"}
	if len(cols) != len(want) {
		t.Fatalf("soft sync columns = %v, want %v", cols, want)
	}
	for i := range want {
		if cols[i] != want[i] {
			t.Errorf("soft sync columns = %v, want %v", cols, want)
			break
		}
	}

	var zzz, aaa string
	if err := db.Raw("SELECT zzz, aaa FROM sync_users WHERE id = 1").Row().Scan(&zzz, &aaa); err != nil {
		t.Fatalf("scan extras: %v", err)
	}
	if zzz != "z-data" || aaa != "a-data" {
		t.Errorf("extra data lost after soft sync: zzz=%q aaa=%q", zzz, aaa)
	}
}

func TestSyncTablesMust(t *testing.T) {
	db := newTestDB(t)
	MustSyncTablesSoft(db, &syncUser{})
	MustSyncTablesHard(db, &syncUser{})
	if !db.Migrator().HasTable(&syncUser{}) {
		t.Fatal("expected sync_users table to exist")
	}
}

// syncPost is a second model used to prove syncTables processes every model
// instead of stopping after the first one.
type syncPost struct {
	ID    int `gorm:"primaryKey"`
	Title string
	Body  string
}

func (syncPost) TableName() string { return "sync_posts" }

// TestSyncTablesMultipleModels guards against the regression where the sync
// loop returned after the first model, silently leaving every later table
// unmigrated while still reporting success.
func TestSyncTablesMultipleModels(t *testing.T) {
	for _, tc := range []struct {
		name string
		sync func(*gorm.DB, ...any) error
	}{
		{"soft", SyncTablesSoft},
		{"hard", SyncTablesHard},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := newTestDB(t)

			// Start from the pre-change schema: no Age and no Body columns.
			if err := db.Exec("CREATE TABLE sync_users (id integer PRIMARY KEY, name text)").Error; err != nil {
				t.Fatalf("create sync_users: %v", err)
			}
			if err := db.Exec("CREATE TABLE sync_posts (id integer PRIMARY KEY, title text)").Error; err != nil {
				t.Fatalf("create sync_posts: %v", err)
			}

			if err := tc.sync(db, &syncUser{}, &syncPost{}); err != nil {
				t.Fatalf("sync: %v", err)
			}

			// Both tables must have gained their new column.
			userCols := tableColumns(t, db, "sync_users")
			if !slices.Contains(userCols, "age") {
				t.Errorf("first model not synced: %v", userCols)
			}
			postCols := tableColumns(t, db, "sync_posts")
			if !slices.Contains(postCols, "body") {
				t.Errorf("second model not synced: %v", postCols)
			}
		})
	}
}

// TestSQLDriverMapping covers the GORM-name to database/sql-name translation and
// the "file database needs no CREATE DATABASE" classification.
func TestSQLDriverMapping(t *testing.T) {
	for _, tc := range []struct {
		gorm       string
		sql        string
		serverBase bool
	}{
		{"mysql", "mysql", true},
		{"mariadb", "mysql", true},
		{"postgres", "pgx", true},
		{"postgresql", "pgx", true},
		{"sqlserver", "sqlserver", true},
		{"mssql", "sqlserver", true},
		{"clickhouse", "clickhouse", true},
		{"sqlite", "sqlite3", false},
		{"sqlite3", "sqlite3", false},
	} {
		drv, serverBased, err := sqlDriverOf(tc.gorm)
		if err != nil {
			t.Errorf("sqlDriverOf(%q): %v", tc.gorm, err)
			continue
		}
		if drv != tc.sql || serverBased != tc.serverBase {
			t.Errorf("sqlDriverOf(%q) = (%q, %v), want (%q, %v)",
				tc.gorm, drv, serverBased, tc.sql, tc.serverBase)
		}
	}

	if _, _, err := sqlDriverOf("oracle"); err == nil {
		t.Error("sqlDriverOf(oracle) should report an unsupported driver")
	}
}

// TestAutoCreateSQLiteIsNoop ensures a file database — which the driver creates
// on demand — does not fail by trying to run CREATE DATABASE.
func TestAutoCreateSQLiteIsNoop(t *testing.T) {
	if err := AutoCreateE("ignored", "sqlite", filepath.Join(t.TempDir(), "x.db")); err != nil {
		t.Fatalf("AutoCreateE(sqlite) = %v, want nil", err)
	}
}

// TestDeleteRejectsRawString guards the SQL-injection footgun: a route parameter
// such as "1 or 1=1" must never reach GORM as a raw WHERE clause.
func TestDeleteRejectsRawString(t *testing.T) {
	db := newTestDB(t)
	if err := db.AutoMigrate(&syncUser{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	for _, name := range []string{"alice", "bob"} {
		if err := db.Create(&syncUser{Name: name}).Error; err != nil {
			t.Fatalf("insert %s: %v", name, err)
		}
	}

	for _, cond := range []string{"1 or 1=1", "1=1", "", "id = 1", "true", "NULL", "now()", "id"} {
		if err := Delete[syncUser](db, cond); err == nil {
			t.Errorf("Delete(%q) accepted a raw condition", cond)
		}
	}
	if n := Count[syncUser](db); n != 2 {
		t.Fatalf("rows after rejected deletes = %d, want 2", n)
	}

	// A digit-only string is GORM's primary-key shorthand and stays allowed.
	if err := Delete[syncUser](db, "1"); err != nil {
		t.Fatalf("Delete(\"1\"): %v", err)
	}
	if n := Count[syncUser](db); n != 1 {
		t.Errorf("rows after Delete(\"1\") = %d, want 1", n)
	}
	if got := IFirst[syncUser](db, 2); got.Name != "bob" {
		t.Errorf("wrong row deleted: %+v", got)
	}

	// The typed helper deletes exactly one row.
	affected, err := DeleteByID[syncUser](db, 2)
	if err != nil {
		t.Fatalf("DeleteByID: %v", err)
	}
	if affected != 1 {
		t.Errorf("DeleteByID affected = %d, want 1", affected)
	}
	if n := Count[syncUser](db); n != 0 {
		t.Errorf("rows after DeleteByID = %d, want 0", n)
	}
}

// TestPaginateWithoutModel guards against calling Paginate on a raw *gorm.DB:
// without a model GORM parsed *int64 as the model and emitted an untabled
// "SELECT count(*)", failing the whole call.
func TestPaginateWithoutModel(t *testing.T) {
	db := newTestDB(t)
	_ = db.AutoMigrate(&user{})
	seedUsers(t, db, []string{"a", "b", "c", "d", "e"})

	page, out := Paginate[user](db, Pagination{Page: 1, Size: 2})
	if out.Error != nil {
		t.Fatalf("Paginate: %v", out.Error)
	}
	if page.Total != 5 {
		t.Errorf("Total = %d, want 5", page.Total)
	}
	if len(page.Content) != 2 {
		t.Errorf("Content len = %d, want 2", len(page.Content))
	}
}

// TestJSONColumnTypeSQLite proves List/Dict declare their own column type, so
// AutoMigrate works on every dialect without a hand-written `type:` tag.
func TestJSONColumnTypeSQLite(t *testing.T) {
	db := newTestDB(t)
	if err := db.AutoMigrate(&user{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	types, err := db.Migrator().ColumnTypes(&user{})
	if err != nil {
		t.Fatalf("column types: %v", err)
	}
	want := map[string]bool{"tags": false, "meta": false}
	for _, col := range types {
		if _, ok := want[col.Name()]; !ok {
			continue
		}
		if got := col.DatabaseTypeName(); got != "text" {
			t.Errorf("column %s type = %q, want text", col.Name(), got)
		}
		want[col.Name()] = true
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("column %s not found", name)
		}
	}
}

func TestDialectorMySQLKeepsParams(t *testing.T) {
	d, err := Dialector("mysql", "u:p@tcp(h:1)/?parseTime=true", "app", "charset=utf8mb4")
	if err != nil {
		t.Fatalf("Dialector: %v", err)
	}
	dsn := d.(*mysql.Dialector).DSN
	// The database must be injected between the path slash and the query, not
	// appended after "?parseTime=...".
	if !strings.Contains(dsn, "/app?") {
		t.Errorf("dsn = %q, want database app before the query", dsn)
	}
	if !strings.Contains(dsn, "parseTime=true") || !strings.Contains(dsn, "charset=utf8mb4") {
		t.Errorf("dsn = %q, lost parameters", dsn)
	}
}

func TestDialectorSQLServerMergesQuery(t *testing.T) {
	d, err := Dialector("sqlserver", "sqlserver://u:p@h:1?encrypt=disable", "app", "TrustServerCertificate=true")
	if err != nil {
		t.Fatalf("Dialector: %v", err)
	}
	raw := d.(*sqlserver.Dialector).DSN
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("url.Parse(%q): %v", raw, err)
	}
	q := u.Query()
	if got := q["database"]; len(got) != 1 || got[0] != "app" {
		t.Errorf("database = %v, want exactly [app]", got)
	}
	if q.Get("encrypt") != "disable" || q.Get("TrustServerCertificate") != "true" {
		t.Errorf("merged query = %v, want encrypt and TrustServerCertificate", q)
	}
}

func TestMaintenanceDSN(t *testing.T) {
	got, err := maintenanceDSN("postgres", "postgres://u:p@h:1/app")
	if err != nil {
		t.Fatalf("maintenanceDSN(url): %v", err)
	}
	if got != "postgres://u:p@h:1/postgres" {
		t.Errorf("maintenanceDSN(url) = %q", got)
	}

	got, err = maintenanceDSN("postgres", "host=h user=u dbname=app port=1")
	if err != nil {
		t.Fatalf("maintenanceDSN(kv): %v", err)
	}
	if strings.Contains(got, "dbname=app") || !strings.Contains(got, "dbname=postgres") {
		t.Errorf("maintenanceDSN(kv) = %q, want dbname=postgres only", got)
	}

	got, err = maintenanceDSN("sqlserver", "sqlserver://u:p@h:1?database=app")
	if err != nil {
		t.Fatalf("maintenanceDSN(sqlserver): %v", err)
	}
	if !strings.Contains(got, "database=master") {
		t.Errorf("maintenanceDSN(sqlserver) = %q, want database=master", got)
	}
}

type syncParent struct {
	ID   int `gorm:"primaryKey"`
	Name string
	Note string
}

func (syncParent) TableName() string { return "parent" }

// sqlRecorder records every statement GORM runs so a no-op sync can be proven
// to emit no DDL.
type sqlRecorder struct {
	logger.Interface
	sqls []string
}

func (r *sqlRecorder) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, _ := fc()
	r.sqls = append(r.sqls, sql)
}

func (r *sqlRecorder) ddlCount() int {
	n := 0
	for _, sql := range r.sqls {
		upper := strings.ToUpper(strings.TrimSpace(sql))
		if strings.HasPrefix(upper, "CREATE") || strings.HasPrefix(upper, "ALTER") ||
			strings.HasPrefix(upper, "DROP") || strings.HasPrefix(upper, "INSERT") {
			n++
		}
	}
	return n
}

// TestSyncTablesSoftPreservesIndexesAndForeignKeys is the core SQLite
// regression: reordering columns used to rebuild via rename, which broke
// indexes and rewrote other tables' foreign keys, and every call rebuilt the
// table even when the order already matched.
func TestSyncTablesSoftPreservesIndexesAndForeignKeys(t *testing.T) {
	db := newTestDB(t)
	for _, stmt := range []string{
		`CREATE TABLE parent (id INTEGER PRIMARY KEY, legacy TEXT, name TEXT)`,
		`CREATE TABLE child (id INTEGER PRIMARY KEY, pid INTEGER REFERENCES parent(id))`,
		`CREATE INDEX idx_parent_name ON parent(name)`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("setup %q: %v", stmt, err)
		}
	}
	if err := db.Exec(`INSERT INTO parent (id, legacy, name) VALUES (1, 'l1', 'p1'), (2, 'l2', 'p2')`).Error; err != nil {
		t.Fatalf("seed parent: %v", err)
	}
	if err := db.Exec(`INSERT INTO child (id, pid) VALUES (10, 1)`).Error; err != nil {
		t.Fatalf("seed child: %v", err)
	}

	if err := SyncTablesSoft(db, &syncParent{}); err != nil {
		t.Fatalf("SyncTablesSoft: %v", err)
	}

	want := []string{"id", "name", "note", "legacy"}
	if cols := tableColumns(t, db, "parent"); !slices.Equal(cols, want) {
		t.Fatalf("columns = %v, want %v", cols, want)
	}

	var childSQL string
	if err := db.Raw("SELECT sql FROM sqlite_master WHERE type='table' AND name='child'").Row().Scan(&childSQL); err != nil {
		t.Fatalf("read child ddl: %v", err)
	}
	if !strings.Contains(childSQL, "parent") || strings.Contains(childSQL, "__sync_new") {
		t.Errorf("child foreign key was rewritten: %s", childSQL)
	}

	var idxCount int64
	if err := db.Raw("SELECT count(*) FROM sqlite_master WHERE type='index' AND name='idx_parent_name'").Scan(&idxCount).Error; err != nil {
		t.Fatalf("read index: %v", err)
	}
	if idxCount != 1 {
		t.Errorf("idx_parent_name missing after rebuild")
	}

	rows, err := db.Raw("PRAGMA foreign_key_check").Rows()
	if err != nil {
		t.Fatalf("foreign_key_check: %v", err)
	}
	if rows.Next() {
		rows.Close()
		t.Error("foreign_key_check reported a violation")
	} else {
		rows.Close()
	}

	var name, legacy string
	if err := db.Raw("SELECT name, legacy FROM parent WHERE id = 2").Row().Scan(&name, &legacy); err != nil {
		t.Fatalf("read row: %v", err)
	}
	if name != "p2" || legacy != "l2" {
		t.Errorf("data lost after rebuild: name=%q legacy=%q", name, legacy)
	}

	// A second sync with the same schema must not touch the table at all.
	rec := &sqlRecorder{Interface: logger.Discard}
	db.Logger = rec
	if err := SyncTablesSoft(db, &syncParent{}); err != nil {
		t.Fatalf("second SyncTablesSoft: %v", err)
	}
	if n := rec.ddlCount(); n != 0 {
		t.Errorf("second sync emitted %d DDL statements: %v", n, rec.sqls)
	}
}

// TestSyncTablesSoftCleansStaleTempTable proves a leftover temporary table from
// an interrupted run cannot poison the next one or lose the original data.
func TestSyncTablesSoftCleansStaleTempTable(t *testing.T) {
	db := newTestDB(t)
	// "junk" before "name" forces a real reorder, so the rebuild path runs.
	if err := db.Exec(`CREATE TABLE parent (id INTEGER PRIMARY KEY, junk TEXT, name TEXT)`).Error; err != nil {
		t.Fatalf("setup parent: %v", err)
	}
	if err := db.Exec(`INSERT INTO parent (id, junk, name) VALUES (1, 'j', 'keep')`).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := db.Exec(`CREATE TABLE parent__sync_new (id INTEGER PRIMARY KEY, garbage TEXT)`).Error; err != nil {
		t.Fatalf("stale temp: %v", err)
	}
	if err := db.Exec(`INSERT INTO parent__sync_new (id, garbage) VALUES (99, 'junk')`).Error; err != nil {
		t.Fatalf("stale row: %v", err)
	}

	if err := SyncTablesSoft(db, &syncParent{}); err != nil {
		t.Fatalf("SyncTablesSoft: %v", err)
	}

	var name string
	if err := db.Raw("SELECT name FROM parent WHERE id = 1").Row().Scan(&name); err != nil {
		t.Fatalf("read parent: %v", err)
	}
	if name != "keep" {
		t.Errorf("parent data damaged: %q", name)
	}
	if db.Migrator().HasTable("parent__sync_new") {
		t.Error("temporary table survived the sync")
	}
}

// TestResetTables drops and recreates the table on one connection, wiping data
// and restoring the struct column order.
func TestResetTables(t *testing.T) {
	db := newTestDB(t)
	if err := db.AutoMigrate(&syncUser{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Create(&syncUser{Name: "alice", Age: 1}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := ResetTables(db, &syncUser{}); err != nil {
		t.Fatalf("ResetTables: %v", err)
	}
	if n := Count[syncUser](db); n != 0 {
		t.Errorf("rows after reset = %d, want 0", n)
	}
	want := []string{"id", "name", "age"}
	if cols := tableColumns(t, db, "sync_users"); !slices.Equal(cols, want) {
		t.Errorf("columns = %v, want %v", cols, want)
	}
}
