package structs

import "testing"

type profile struct {
	Name string `mapstructure:"name"`
	Age  int    `mapstructure:"age"`
	OK   bool   `mapstructure:"ok"`
}

func TestMap(t *testing.T) {
	m := Map(profile{Name: "alice", Age: 30, OK: true})
	if m["name"] != "alice" || m["age"] != 30 || m["ok"] != true {
		t.Errorf("Map = %v, want name/alice age/30 ok/true", m)
	}
}

func TestTo(t *testing.T) {
	p := To[profile](map[string]any{"name": "bob", "age": 25, "ok": false})
	if p.Name != "bob" || p.Age != 25 || p.OK {
		t.Errorf("To = %+v, want bob/25/false", p)
	}
}

func TestToFromStruct(t *testing.T) {
	src := profile{Name: "carol", Age: 40, OK: true}
	dst := To[profile](src)
	if dst != src {
		t.Errorf("To(struct) = %+v, want %+v", dst, src)
	}
}

func TestVoValid(t *testing.T) {
	v, err := Vo[struct {
		Name string `mapstructure:"name" validate:"required"`
	}](map[string]any{"name": "dave"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if v.Name != "dave" {
		t.Errorf("Name = %q, want dave", v.Name)
	}
}

func TestVoInvalid(t *testing.T) {
	if _, err := Vo[struct {
		Name string `mapstructure:"name" validate:"required"`
	}](map[string]any{}); err == nil {
		t.Error("expected validation error for missing name")
	}
}