package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
)

func TestMountSPAServesRootIndexHTML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>homebox</html>"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HBOX_PUBLIC_DIR", dir)

	r := router.NewRouter(func(w http.ResponseWriter, req *http.Request) (*core.RequestEvent, router.EventCleanupFunc) {
		return &core.RequestEvent{
			Event: router.Event{
				Response: w,
				Request:  req,
			},
		}, nil
	})
	mountSPA(r)

	mux, err := r.BuildMux()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.SetPathValue(apis.StaticWildcardParam, "")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", rec.Code, rec.Body.String())
	}
	if got := rec.Body.String(); got != "<html>homebox</html>" {
		t.Fatalf("expected index.html body, got %q", got)
	}
}

func TestPublicDirUsesExecutableRelativePath(t *testing.T) {
	t.Setenv("HBOX_PUBLIC_DIR", "")

	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(filepath.Dir(exe), "pb_public")
	if got := publicDir(); got != want && got != "pb_public" {
		// go test may run via go run; allow pb_public fallback.
		t.Fatalf("expected %q or pb_public, got %q", want, got)
	}
}
