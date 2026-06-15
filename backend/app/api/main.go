package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/hay-kot/homebox/backend/internal/core/currencies"
	homeboxpb "github.com/hay-kot/homebox/backend/internal/pocketbase"
	"github.com/hay-kot/homebox/backend/internal/sys/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

var (
	version   = "nightly"
	commit    = "HEAD"
	buildTime = "now"
)

func build() string {
	short := commit
	if len(short) > 7 {
		short = short[:7]
	}
	return fmt.Sprintf("%s, commit %s, built at %s", version, short, buildTime)
}

func main() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	cfg, err := config.New(build(), "Homebox inventory management system")
	if err != nil {
		panic(err)
	}

	if err := run(cfg); err != nil {
		panic(err)
	}
}

func run(cfg *config.Config) error {
	setupLogger(cfg)

	if err := os.MkdirAll(cfg.Storage.Data, 0o755); err != nil {
		log.Fatal().Err(err).Msg("failed to create data directory")
	}
	if err := os.MkdirAll(cfg.Storage.PocketBaseDataDir(), 0o755); err != nil {
		log.Fatal().Err(err).Msg("failed to create pocketbase data directory")
	}

	currencyList, err := loadCurrencies(cfg)
	if err != nil {
		return err
	}

	app := homeboxpb.New(homeboxpb.Options{
		Config:    cfg,
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
	}, currencyList)

	registerSPA(app)

	httpAddr := cfg.Web.Host + ":" + cfg.Web.Port
	if cfg.Web.Host == "" {
		httpAddr = "0.0.0.0:" + cfg.Web.Port
	}
	os.Args = []string{os.Args[0], "serve", "--http=" + httpAddr}

	return app.Start()
}

func loadCurrencies(cfg *config.Config) ([]currencies.Currency, error) {
	collectFuncs := []currencies.CollectorFunc{currencies.CollectDefaults()}
	if cfg.Options.CurrencyConfig != "" {
		content, err := os.ReadFile(cfg.Options.CurrencyConfig)
		if err != nil {
			return nil, err
		}
		collectFuncs = append(collectFuncs, currencies.CollectJSON(bytes.NewReader(content)))
	}
	return currencies.CollectionCurrencies(collectFuncs...)
}

func setupLogger(cfg *config.Config) {
	if cfg.Log.Format != config.LogFormatJSON {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()
	}
	level, err := zerolog.ParseLevel(cfg.Log.Level)
	if err == nil {
		zerolog.SetGlobalLevel(level)
	}
}
