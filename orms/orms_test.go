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

func TestDsn(t *testing.T) {
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
		got, err := Dsn(c.driver, "db", "u", "p", "h", "1")
		if err != nil {
			t.Errorf("Dsn(%s) error: %v", c.driver, err)
			continue
		}
		if got != c.want {
			t.Errorf("Dsn(%s) = %q, want %q", c.driver, got, c.want)
		}
	}
	if _, err := Dsn("unknown", "db", "u", "p", "h", "1"); err == nil {
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