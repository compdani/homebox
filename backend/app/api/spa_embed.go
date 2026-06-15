package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"
	"github.com/pocketbase/pocketbase/tools/osutils"
	"github.com/pocketbase/pocketbase/tools/router"
)

func publicDir() string {
	if dir := os.Getenv("HBOX_PUBLIC_DIR"); dir != "" {
		return dir
	}
	if osutils.IsProbablyGoRun() {
		return "pb_public"
	}

	exe, err := os.Executable()
	if err != nil {
		return "pb_public"
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		exe, _ = os.Executable()
	}
	return filepath.Join(filepath.Dir(exe), "pb_public")
}

func mountSPA(r *router.Router[*core.RequestEvent]) {
	dir := publicDir()
	info, err := os.Stat(dir)
	if err != nil {
		log.Printf("homebox: SPA static dir unavailable (%s): %v", dir, err)
		return
	}
	if !info.IsDir() {
		log.Printf("homebox: SPA static path is not a directory: %s", dir)
		return
	}

	if r.HasRoute(http.MethodGet, "/{path...}") {
		return
	}

	r.GET("/{path...}", apis.Static(os.DirFS(dir), true))
}

func registerSPA(app core.App) {
	app.OnServe().Bind(&hook.Handler[*core.ServeEvent]{
		Priority: 999,
		Func: func(se *core.ServeEvent) error {
			mountSPA(se.Router)
			return se.Next()
		},
	})
}
