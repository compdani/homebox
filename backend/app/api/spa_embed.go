package main

import (
	"embed"
	"errors"
	"io"
	"mime"
	"path"
	"path/filepath"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
)

//go:embed all:static/public/*
var public embed.FS

var errDir = errors.New("path is dir")

func mountSPA(r *router.Router[*core.RequestEvent]) {
	registerSPAMimes()
	r.GET("/{path...}", func(e *core.RequestEvent) error {
		if tryServeSPA(e, e.Request.URL.Path) == nil {
			return nil
		}
		return tryServeSPA(e, "/index.html")
	})
}

func tryServeSPA(e *core.RequestEvent, requestedPath string) error {
	f, err := public.Open(path.Join("static/public", requestedPath))
	if err != nil {
		return err
	}
	defer f.Close()
	stat, _ := f.Stat()
	if stat.IsDir() {
		return errDir
	}
	contentType := mime.TypeByExtension(filepath.Ext(requestedPath))
	e.Response.Header().Set("Content-Type", contentType)
	_, err = io.Copy(e.Response, f)
	return err
}

func registerSPAMimes() {
	_ = mime.AddExtensionType(".js", "application/javascript")
	_ = mime.AddExtensionType(".mjs", "application/javascript")
}
