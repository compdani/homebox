package main

import (
	"errors"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
)

var errDir = errors.New("path is dir")

func publicDir() string {
	if dir := os.Getenv("HBOX_PUBLIC_DIR"); dir != "" {
		return dir
	}
	return "pb_public"
}

func mountSPA(r *router.Router[*core.RequestEvent]) {
	dir := publicDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return
	}

	registerSPAMimes()
	serve := func(e *core.RequestEvent) error {
		requestedPath := e.Request.URL.Path
		if strings.HasPrefix(requestedPath, "/api/") || strings.HasPrefix(requestedPath, "/_") {
			return os.ErrNotExist
		}
		if tryServeSPAFromDisk(e, dir, requestedPath) == nil {
			return nil
		}
		return tryServeSPAFromDisk(e, dir, "/index.html")
	}
	// PocketBase auto-registers a JSON 404 for "/" unless HasRoute("", "/") is true.
	// Use a method-agnostic "/" route (not GET /) so it doesn't conflict with GET /{path...}.
	r.Route("", "/", serve)
	r.GET("/{path...}", serve)
}

func tryServeSPAFromDisk(e *core.RequestEvent, dir, requestedPath string) error {
	clean := filepath.Clean(requestedPath)
	if strings.Contains(clean, "..") {
		return os.ErrNotExist
	}

	full := filepath.Join(dir, filepath.FromSlash(strings.TrimPrefix(clean, "/")))
	info, err := os.Stat(full)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errDir
	}

	f, err := os.Open(full)
	if err != nil {
		return err
	}
	defer f.Close()

	contentType := mime.TypeByExtension(filepath.Ext(full))
	e.Response.Header().Set("Content-Type", contentType)
	_, err = io.Copy(e.Response, f)
	return err
}

func registerSPAMimes() {
	_ = mime.AddExtensionType(".html", "text/html; charset=utf-8")
	_ = mime.AddExtensionType(".js", "application/javascript")
	_ = mime.AddExtensionType(".mjs", "application/javascript")
	_ = mime.AddExtensionType(".webmanifest", "application/manifest+json")
	_ = mime.AddExtensionType(".json", "application/json")
}
