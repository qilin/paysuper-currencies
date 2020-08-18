package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/handlers"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/service/grpc"
	"github.com/micro/go-plugins/client/selector/static"
	"github.com/micro/go-plugins/wrapper/monitoring/prometheus"
	"github.com/paysuper/paysuper-currencies/config"
	"github.com/paysuper/paysuper-currencies/internal/service"
	"github.com/paysuper/paysuper-database-mongo"
	currencies "github.com/paysuper/paysuper-proto/go/currenciespb"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal("Config init failed with error", zap.Error(err))
	}

	migrations, err := migrate.New("file://./migrations", cfg.MongoDsn)

	if err != nil {
		logger.Fatal("Migrations initialization failed", zap.Error(err))
	}

	err = migrations.Up()

	if err != nil && err != migrate.ErrNoChange && err != migrate.ErrNilVersion {
		logger.Fatal("Migrations processing failed", zap.Error(err))
	}

	logger.Info("db migrations applied")

	db, err := database.NewDatabase()
	if err != nil {
		logger.Fatal("Database connection failed", zap.Error(err))
	}

	cs, err := service.NewService(cfg, db)
	if err != nil {
		logger.Fatal("Can`t create currency rates service", zap.Error(err))
	}

	var source string
	flag.StringVar(&source, "source", "", "rates source")
	flag.Parse()

	if source != "" {
		logger.Info("Updating currency rates from " + source)

		defer db.Close()

		g := errgroup.Group{}

		var err error

		switch source {
		case "oxr":
			err = cs.RequestRatesOxr()
			if err == nil {
				g.Go(func() error {
					return cs.SetRatesPaysuper()
				})
				g.Go(func() error {
					return cs.SetRatesStock()
				})
			}
		case "paysuper":
			g.Go(func() error {
				return cs.SetRatesPaysuper()
			})
		case "centralbanks":
			g.Go(func() error {
				return cs.RequestRatesCbrf()
			})
			g.Go(func() error {
				return cs.RequestRatesCbeu()
			})
			g.Go(func() error {
				return cs.RequestRatesCbca()
			})
			g.Go(func() error {
				return cs.RequestRatesCbpl()
			})
			g.Go(func() error {
				return cs.RequestRatesCbau()
			})
			g.Go(func() error {
				return cs.RequestRatesCbtr()
			})
		case "stock":
			g.Go(func() error {
				return cs.SetRatesStock()
			})
		default:
			logger.Fatal("Source is unknown, exiting")
		}

		if err == nil {
			err = g.Wait()
		}

		if err != nil {
			logger.Fatal("Updating currency rates error", zap.Error(err))
		}

		return
	}

	err = cs.Init()
	if err != nil {
		logger.Fatal("Service init failed", zap.Error(err))
	}

	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.Handler())

	h := health.New()
	err = h.AddChecks([]*health.Config{
		{
			Name:     "health-check",
			Checker:  cs,
			Interval: time.Duration(1) * time.Second,
			Fatal:    true,
		},
	})
	if err != nil {
		logger.Fatal("Health check register failed", zap.Error(err))
	}
	router.HandleFunc("/_healthz", handlers.NewJSONHandlerFunc(h, nil))

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.MetricsPort),
		Handler: router,
	}

	var ms micro.Service
	options := []micro.Option{
		micro.Name(currencies.ServiceName),
		micro.Version(currencies.Version),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.BeforeStart(func() error {
			go func() {
				logger.Info("Metrics and health check listening", zap.Int("port", cfg.MetricsPort))
				if err = httpServer.ListenAndServe(); err != nil {
					logger.Error("Metrics and health check listen failed", zap.Error(err))
				}
			}()
			return nil
		}),
		micro.AfterStop(func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := httpServer.Shutdown(ctx); err != nil {
				logger.Fatal("Http server shutdown failed", zap.Error(err))
			}
			logger.Info("Http server stopped")

			db.Close()
			logger.Info("Db closed")

			if err := logger.Sync(); err != nil {
				logger.Fatal("Logger sync failed", zap.Error(err))
			} else {
				logger.Info("Logger synced")
			}
			return nil
		}),
	}

	if cfg.MicroSelector == "static" {
		zap.L().Info(`Use micro selector "static"`)
		options = append(options, micro.Selector(static.NewSelector()))
	}

	logger.Info("Initialize micro service")

	// todo: micro.NewService replaced by grpc.NewService to get native gRPC client support
	//ms = micro.NewService(options...)
	ms = grpc.NewService(options...)
	ms.Init()

	err = currencies.RegisterCurrencyRatesServiceHandler(ms.Server(), cs)
	if err != nil {
		logger.Fatal("Can`t register service in micro", zap.Error(err))
	}

	if err := ms.Run(); err != nil {
		logger.Fatal("Can`t run service", zap.Error(err))
	}
}
