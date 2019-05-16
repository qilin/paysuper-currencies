package main

import (
    "flag"
    "fmt"
    "github.com/InVisionApp/go-health"
    "github.com/InVisionApp/go-health/handlers"
    "github.com/micro/go-micro"
    "github.com/micro/go-plugins/wrapper/monitoring/prometheus"
    k8s "github.com/micro/kubernetes/go/micro"
    "github.com/paysuper/paysuper-currencies-rates/config"
    "github.com/paysuper/paysuper-currencies-rates/internal"
    "github.com/paysuper/paysuper-currencies-rates/pkg"
    "github.com/paysuper/paysuper-currencies-rates/pkg/proto/currencyrates"
    "github.com/paysuper/paysuper-database-mongo"
    "github.com/paysuper/paysuper-recurring-repository/pkg/constant"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "go.uber.org/zap"
    "net/http"
    "time"
)

func main() {
    logger, _ := zap.NewProduction()
    zap.ReplaceGlobals(logger)

    cfg, err := config.NewConfig()
    if err != nil {
        logger.Fatal("Config init failed with error", zap.Error(err))
    }

    settings := database.Connection{
        Host:     cfg.MongoHost,
        Database: cfg.MongoDatabase,
        User:     cfg.MongoUser,
        Password: cfg.MongoPassword,
    }
    db, err := database.NewDatabase(settings)
    if err != nil {
        logger.Fatal("Database connection failed", zap.Error(err), zap.String("connection_string", settings.String()))
    }

    cs, err := internal.NewService(cfg, db)
    if err != nil {
        logger.Fatal("Can`t create currency rates service", zap.Error(err))
    }

    var source string
    flag.StringVar(&source, "source", "", "rates source")
    flag.Parse()

    if source != "" {
        logger.Info("Updating currency rates from " + source)

        defer db.Close()

        var err error

        switch source {
        case "oxr":
            err = cs.RequestRatesOxr()
            if err == nil {
                err = cs.SetRatesPaysuper()
            }
            if err == nil {
                err = cs.SetRatesStock()
            }
        case "paysuper":
            err = cs.SetRatesPaysuper()
        case "centralbanks":
            err = cs.RequestRatesCbrf()
            if err == nil {
                err = cs.RequestRatesCbeu()
            }
            if err == nil {
                err = cs.RequestRatesCbca()
            }
            if err == nil {
                err = cs.RequestRatesCbpl()
            }
            if err == nil {
                err = cs.RequestRatesCbau()
            }
        case "stock":
            err = cs.SetRatesStock()
        case "cardpay":
            err = cs.RequestRatesCardpay()
            if err == nil {
                err = cs.CalculatePaysuperCorrections()
            }
        default:
            logger.Fatal("Source is unknown, exiting")
        }

        if err != nil {
            logger.Fatal("Updating currency rates error", zap.Error(err))
        }

        return
    }

    initHealth(cs, cfg.MetricsPort)
    initPrometheus()

    var ms micro.Service
    options := []micro.Option{
        micro.Name(pkg.ServiceName),
        micro.Version(pkg.Version),
        micro.WrapHandler(prometheus.NewHandlerWrapper()),
        micro.BeforeStart(func() error {
            go func() {
                if err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.MetricsPort), nil); err != nil {
                    logger.Fatal("Metrics listen failed", zap.Error(err))
                }
            }()
            return nil
        }),
        micro.AfterStop(func() error {
            db.Close()
            return nil
        }),
    }

    if cfg.MicroRegistry == constant.RegistryKubernetes {
        ms = k8s.NewService(options...)
        logger.Info("Initialize k8s service")
    } else {
        ms = micro.NewService(options...)
        logger.Info("Initialize micro service")
    }

    ms.Init()

    err = currencyrates.RegisterCurrencyratesServiceHandler(ms.Server(), cs)
    if err != nil {
        logger.Fatal("Can`t register service in micro", zap.Error(err))
    }

    if err := ms.Run(); err != nil {
        logger.Fatal("Can`t run service", zap.Error(err))
    }
}

func initHealth(checker health.ICheckable, port int) {
    h := health.New()
    err := h.AddChecks([]*health.Config{
        {
            Name:     "health-check",
            Checker:  checker,
            Interval: time.Duration(1) * time.Second,
            Fatal:    true,
        },
    })

    if err != nil {
        zap.L().Fatal("Health check register failed", zap.Error(err))
    }

    zap.L().Info("Health check listening port", zap.Int("port", port))

    if err = h.Start(); err != nil {
        zap.L().Fatal("Health check start failed", zap.Error(err))
    }

    http.HandleFunc("/health", handlers.NewJSONHandlerFunc(h, nil))
}

func initPrometheus() {
    http.Handle("/metrics", promhttp.Handler())
}
