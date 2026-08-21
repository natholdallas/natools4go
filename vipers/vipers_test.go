package vipers

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func setup(t *testing.T) {
	t.Helper()
	viper.Reset()
}

func TestConfigReadsTOML(t *testing.T) {
	setup(t)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.toml"), []byte(`
[server]
port = 9090
debug = true
name = "demo"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	Config("app", dir, "toml")

	if got := Int("server.port"); got != 9090 {
		t.Errorf("port = %d, want 9090", got)
	}
	if got := Bool("server.debug"); !got {
		t.Error("debug should be true")
	}
	if got := String("server.name"); got != "demo" {
		t.Errorf("name = %q, want demo", got)
	}
}

func TestGettersUseDefaults(t *testing.T) {
	setup(t)
	if got := Int("missing.int", 7); got != 7 {
		t.Errorf("Int default = %d, want 7", got)
	}
	if got := String("missing.str", "fallback"); got != "fallback" {
		t.Errorf("String default = %q, want fallback", got)
	}
	if got := Bool("missing.bool", true); !got {
		t.Error("Bool default should be true")
	}
	if got := Duration("missing.dur", 5*time.Second); got != 5*time.Second {
		t.Errorf("Duration default = %v, want 5s", got)
	}
}

func TestGetGeneric(t *testing.T) {
	setup(t)
	if got := Get("missing", 3); got != 3 {
		t.Errorf("Get = %d, want 3", got)
	}
	if got := Get("missing", "x"); got != "x" {
		t.Errorf("Get string = %q, want x", got)
	}
}

func TestReloadDispatches(t *testing.T) {
	setup(t)
	var calls atomic.Int32
	NewUpdateEvent(func(e fsnotify.Event) { calls.Add(1) })

	Reload(fsnotify.Event{Name: "app.toml"})
	Reload(fsnotify.Event{Name: "app.toml"})

	if got := calls.Load(); got != 2 {
		t.Errorf("handlers called %d times, want 2", got)
	}
}

func TestNewUpdateEventSkipsNil(t *testing.T) {
	setup(t)
	NewUpdateEvent(nil)
	Reload(fsnotify.Event{})
	// No panic is the assertion.
}
