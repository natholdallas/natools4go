package orms

import (
	"io"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
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
		{"mysql", "u:p@tcp(h:1)/"},
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
	ID   int    `gorm:"primaryKey"`
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
