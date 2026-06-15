package homeboxpb

import (
	"path/filepath"

	customapi "github.com/hay-kot/homebox/backend/app/api/custom"
	"github.com/hay-kot/homebox/backend/internal/core/currencies"
	"github.com/hay-kot/homebox/backend/internal/pocketbase/hooks"
	"github.com/hay-kot/homebox/backend/internal/pocketbase/store"
	"github.com/hay-kot/homebox/backend/internal/sys/config"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	_ "github.com/hay-kot/homebox/backend/internal/pocketbase/migrations"
)

// Options holds runtime configuration for the embedded PocketBase app.
type Options struct {
	Config    *config.Config
	Version   string
	Commit    string
	BuildTime string
}

// New creates a configured PocketBase application.
func New(opts Options, currencyList []currencies.Currency) *pocketbase.PocketBase {
	dataDir := opts.Config.Storage.PocketBaseDataDir()
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: dataDir,
	})

	migrationsDir := filepath.Join(opts.Config.Storage.Data, "migrations")
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Dir:         migrationsDir,
		Automigrate: false,
	})

	st := store.New(app)
	hooks.Register(app)

	customapi.Register(app, customapi.Deps{
		Store:      st,
		Config:     opts.Config,
		Version:    opts.Version,
		Commit:     opts.Commit,
		BuildTime:  opts.BuildTime,
		Currencies: currencyList,
	})

	return app
}

// AppStore returns a store for the bootstrapped app.
func AppStore(app core.App) *store.Store {
	return store.New(app)
}
