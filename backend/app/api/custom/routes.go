package custom

import (
	"context"
	"net/http"
	"time"

	"github.com/hay-kot/homebox/backend/internal/core/currencies"
	"github.com/hay-kot/homebox/backend/internal/pocketbase/store"
	"github.com/hay-kot/homebox/backend/internal/sys/config"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
)

// Deps holds dependencies for custom API routes.
type Deps struct {
	Store      *store.Store
	Config     *config.Config
	Version    string
	Commit     string
	BuildTime  string
	MountSPA   func(*router.Router[*core.RequestEvent])
	Currencies []currencies.Currency
}

// Register mounts Homebox custom /api/v1 routes on the PocketBase router.
func Register(app core.App, deps Deps) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		g := se.Router.Group("/api/v1")

		g.GET("/status", func(e *core.RequestEvent) error {
			return e.JSON(http.StatusOK, store.APISummary{
				Health: true,
				Build: store.Build{
					Version:   deps.Version,
					Commit:    deps.Commit,
					BuildTime: deps.BuildTime,
				},
				Versions:          []string{deps.Version, deps.Commit, deps.BuildTime},
				Title:             "Homebox",
				Message:           "Track, Manage, and Organize your Things.",
				AllowRegistration: deps.Config.Options.AllowRegistration,
				Demo:              deps.Config.Demo,
			})
		})

		g.GET("/currencies", func(e *core.RequestEvent) error {
			return e.JSON(http.StatusOK, deps.Currencies)
		})

		g.POST("/users/register", func(e *core.RequestEvent) error {
			if !deps.Config.Options.AllowRegistration {
				return e.ForbiddenError("registration disabled", nil)
			}
			var body store.UserRegistration
			if err := e.BindBody(&body); err != nil {
				return e.BadRequestError("invalid body", err)
			}
			user, err := deps.Store.RegisterUser(e.Request.Context(), body)
			if err != nil {
				return e.BadRequestError("registration failed", err)
			}
			return apis.RecordAuthResponse(e, user, "password", nil)
		})

		registerQRCodeRoute(g)

		authed := g.Group("")
		authed.Bind(apis.RequireAuth("users"))

		authed.GET("/groups", func(e *core.RequestEvent) error {
			group, err := deps.Store.GroupByID(e.Request.Context(), authGroupID(e))
			if err != nil {
				return e.NotFoundError("group not found", err)
			}
			return e.JSON(http.StatusOK, group)
		})

		authed.PUT("/groups", func(e *core.RequestEvent) error {
			if err := requireOwner(e); err != nil {
				return err
			}
			var body store.GroupUpdate
			if err := e.BindBody(&body); err != nil {
				return e.BadRequestError("invalid body", err)
			}
			group, err := deps.Store.GroupUpdate(e.Request.Context(), authGroupID(e), body)
			if err != nil {
				return e.BadRequestError("update failed", err)
			}
			return e.JSON(http.StatusOK, group)
		})

		authed.POST("/groups/invitations", func(e *core.RequestEvent) error {
			if err := requireOwner(e); err != nil {
				return err
			}
			var body struct {
				Uses      int       `json:"uses"`
				ExpiresAt time.Time `json:"expiresAt"`
			}
			if err := e.BindBody(&body); err != nil {
				return e.BadRequestError("invalid body", err)
			}
			inv, _, err := deps.Store.CreateInvitation(e.Request.Context(), authGroupID(e), body.Uses, body.ExpiresAt)
			if err != nil {
				return e.BadRequestError("invitation failed", err)
			}
			return e.JSON(http.StatusCreated, inv)
		})

		registerStatsRoutes(authed, deps)
		registerActionRoutes(authed, deps)
		registerItemRoutes(authed, deps)
		registerLocationRoutes(authed, deps)
		registerLabelRoutes(authed, deps)
		registerReportRoutes(authed, deps)
		registerNotifierRoutes(authed, deps)

		if deps.MountSPA != nil {
			deps.MountSPA(se.Router)
		}
		return se.Next()
	})

	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		go runCron(deps)
		return nil
	})
}

func authGroupID(e *core.RequestEvent) string {
	if e.Auth == nil {
		return ""
	}
	return e.Auth.GetString("group")
}

func requireOwner(e *core.RequestEvent) error {
	if e.Auth == nil || e.Auth.GetString("role") != "owner" {
		return e.ForbiddenError("owner role required", nil)
	}
	return nil
}

func registerStatsRoutes(g *router.RouterGroup[*core.RequestEvent], deps Deps) {
	g.GET("/groups/statistics", func(e *core.RequestEvent) error {
		stats, err := deps.Store.StatsGroup(e.Request.Context(), authGroupID(e))
		if err != nil {
			return e.InternalServerError("stats failed", err)
		}
		return e.JSON(http.StatusOK, stats)
	})
	g.GET("/groups/statistics/purchase-price", func(e *core.RequestEvent) error {
		start, end := parseDateRange(e)
		stats, err := deps.Store.StatsPurchasePrice(e.Request.Context(), authGroupID(e), start, end)
		if err != nil {
			return e.InternalServerError("stats failed", err)
		}
		return e.JSON(http.StatusOK, stats)
	})
	g.GET("/groups/statistics/locations", func(e *core.RequestEvent) error {
		stats, err := deps.Store.StatsLocationsByPurchasePrice(e.Request.Context(), authGroupID(e))
		if err != nil {
			return e.InternalServerError("stats failed", err)
		}
		return e.JSON(http.StatusOK, stats)
	})
	g.GET("/groups/statistics/labels", func(e *core.RequestEvent) error {
		stats, err := deps.Store.StatsLabelsByPurchasePrice(e.Request.Context(), authGroupID(e))
		if err != nil {
			return e.InternalServerError("stats failed", err)
		}
		return e.JSON(http.StatusOK, stats)
	})
}

func registerActionRoutes(g *router.RouterGroup[*core.RequestEvent], deps Deps) {
	postAction := func(path string, fn func(context.Context, string) (int, error)) {
		g.POST(path, func(e *core.RequestEvent) error {
			if err := requireOwner(e); err != nil {
				return err
			}
			n, err := fn(e.Request.Context(), authGroupID(e))
			if err != nil {
				return e.InternalServerError("action failed", err)
			}
			return e.JSON(http.StatusOK, store.ActionAmountResult{Completed: n})
		})
	}
	postAction("/actions/ensure-asset-ids", deps.Store.EnsureAssetIDs)
	postAction("/actions/ensure-import-refs", deps.Store.EnsureImportRefs)
	postAction("/actions/zero-item-time-fields", deps.Store.ZeroOutTimeFields)
	postAction("/actions/set-primary-photos", deps.Store.SetPrimaryPhotos)
}

func registerItemRoutes(g *router.RouterGroup[*core.RequestEvent], deps Deps) {
	g.GET("/items/{id}/path", func(e *core.RequestEvent) error {
		id := e.Request.PathValue("id")
		itemPath, err := deps.Store.ItemPath(e.Request.Context(), authGroupID(e), id)
		if err != nil {
			return e.NotFoundError("item not found", err)
		}
		return e.JSON(http.StatusOK, itemPath)
	})
	g.POST("/items/place", func(e *core.RequestEvent) error {
		var body store.PlaceItemRequest
		if err := e.BindBody(&body); err != nil {
			return e.BadRequestError("invalid body", err)
		}
		result, err := deps.Store.PlaceItem(e.Request.Context(), authGroupID(e), body)
		if err != nil {
			return e.BadRequestError(err.Error(), err)
		}
		return e.JSON(http.StatusOK, result)
	})
	g.POST("/items/unplace", func(e *core.RequestEvent) error {
		var body store.UnplaceItemRequest
		if err := e.BindBody(&body); err != nil {
			return e.BadRequestError("invalid body", err)
		}
		result, err := deps.Store.UnplaceProduct(e.Request.Context(), authGroupID(e), body)
		if err != nil {
			return e.BadRequestError(err.Error(), err)
		}
		return e.JSON(http.StatusOK, result)
	})
	g.POST("/items/import", func(e *core.RequestEvent) error {
		if err := requireOwner(e); err != nil {
			return err
		}
		file, _, err := e.Request.FormFile("csv")
		if err != nil {
			return e.BadRequestError("missing csv file", err)
		}
		defer file.Close()
		n, err := deps.Store.ImportCSV(e.Request.Context(), authGroupID(e), file, deps.Config.Options.AutoIncrementAssetID)
		if err != nil {
			return e.BadRequestError("import failed", err)
		}
		return e.JSON(http.StatusOK, store.ActionAmountResult{Completed: n})
	})
	g.GET("/items/export", func(e *core.RequestEvent) error {
		rows, err := deps.Store.ExportCSV(e.Request.Context(), authGroupID(e))
		if err != nil {
			return e.InternalServerError("export failed", err)
		}
		w := e.Response
		w.Header().Set("Content-Type", "text/tab-separated-values")
		for _, row := range rows {
			for i, col := range row {
				if i > 0 {
					_, _ = w.Write([]byte("\t"))
				}
				_, _ = w.Write([]byte(col))
			}
			_, _ = w.Write([]byte("\n"))
		}
		return nil
	})
}

func registerLocationRoutes(g *router.RouterGroup[*core.RequestEvent], deps Deps) {
	g.GET("/locations/tree", func(e *core.RequestEvent) error {
		tree, err := deps.Store.LocationTree(e.Request.Context(), authGroupID(e))
		if err != nil {
			return e.InternalServerError("tree failed", err)
		}
		return e.JSON(http.StatusOK, tree)
	})
}

func registerReportRoutes(g *router.RouterGroup[*core.RequestEvent], deps Deps) {
	g.GET("/reporting/bill-of-materials", func(e *core.RequestEvent) error {
		data, err := deps.Store.ExportBillOfMaterials(e.Request.Context(), authGroupID(e))
		if err != nil {
			return e.InternalServerError("export failed", err)
		}
		w := e.Response
		w.Header().Set("Content-Type", "text/tab-separated-values")
		_, _ = w.Write(data)
		return nil
	})
}

func registerNotifierRoutes(g *router.RouterGroup[*core.RequestEvent], deps Deps) {
	g.POST("/notifiers/test", func(e *core.RequestEvent) error {
		var body struct {
			ID string `json:"id"`
		}
		if err := e.BindBody(&body); err != nil {
			return e.BadRequestError("invalid body", err)
		}
		if err := deps.Store.TestNotifier(e.Request.Context(), e.Auth.Id, body.ID); err != nil {
			return e.BadRequestError("test failed", err)
		}
		return e.NoContent(http.StatusOK)
	})
}

func parseDateRange(e *core.RequestEvent) (time.Time, time.Time) {
	parse := func(s string, def time.Time) time.Time {
		if s == "" {
			return def
		}
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return def
		}
		return t
	}
	now := time.Now()
	return parse(e.Request.URL.Query().Get("start"), now.AddDate(0, -1, 0)), parse(e.Request.URL.Query().Get("end"), now)
}

func runCron(deps Deps) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		if time.Now().Hour() == 8 {
			_ = deps.Store.SendNotifiersToday(context.Background())
		}
		_, _ = deps.Store.PurgeInvitations(context.Background())
	}
}
