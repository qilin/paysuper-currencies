package main

import (
    "context"
    "errors"
    "flag"
    "fmt"
    "github.com/InVisionApp/go-health"
    "github.com/InVisionApp/go-health/handlers"
    "github.com/micro/go-micro"
    "github.com/micro/go-plugins/wrapper/monitoring/prometheus"
    k8s "github.com/micro/kubernetes/go/micro"
    "github.com/paysuper/paysuper-currencies/config"
    "github.com/paysuper/paysuper-currencies/internal/service"
    "github.com/paysuper/paysuper-currencies/pkg"
    "github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
    "github.com/paysuper/paysuper-database-mongo"
    "github.com/paysuper/paysuper-recurring-repository/pkg/constant"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "go.uber.org/zap"
    "net/http"
    "time"
)

var updateError = errors.New("Updating currency rates failed")

func main() {
    logger, _ := zap.NewProduction()
    zap.ReplaceGlobals(logger)

    cfg, err := config.NewConfig()
    if err != nil {
        logger.Fatal("Config init failed with error", zap.Error(err))
    }

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

        var err error

        switch source {
        case "oxr":
            err = cs.RequestRatesOxr()
            if err == nil {
                c := make(chan error)
                go cs.SetRatesPaysuper(c)
                go cs.SetRatesStock(c)
                err1, err2 := <-c, <-c
                if err1 != nil {
                    logger.Error("Updating currency rates error", zap.Error(err1))
                    err = updateError
                }
                if err2 != nil {
                    logger.Error("Updating currency rates error", zap.Error(err2))
                    err = updateError
                }
            }
        case "paysuper":
            c := make(chan error)
            go cs.SetRatesPaysuper(c)
            err = <-c
        case "centralbanks":
            c := make(chan error)
            go cs.RequestRatesCbrf(c)
            go cs.RequestRatesCbeu(c)
            go cs.RequestRatesCbca(c)
            go cs.RequestRatesCbpl(c)
            go cs.RequestRatesCbau(c)
            err1, err2, err3, err4, err5 := <-c, <-c, <-c, <-c, <-c
            if err1 != nil {
                logger.Error("Updating currency rates error", zap.Error(err1))
                err = updateError
            }
            if err2 != nil {
                logger.Error("Updating currency rates error", zap.Error(err2))
                err = updateError
            }
            if err3 != nil {
                logger.Error("Updating currency rates error", zap.Error(err3))
                err = updateError
            }
            if err4 != nil {
                logger.Error("Updating currency rates error", zap.Error(err4))
                err = updateError
            }
            if err5 != nil {
                logger.Error("Updating currency rates error", zap.Error(err5))
                err = updateError
            }
        case "stock":
            c := make(chan error)
            go cs.SetRatesStock(c)
            err = <-c
        default:
            logger.Fatal("Source is unknown, exiting")
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
    router.HandleFunc("/health", handlers.NewJSONHandlerFunc(h, nil))

    httpServer := &http.Server{
        Addr:    fmt.Sprintf(":%d", cfg.MetricsPort),
        Handler: router,
    }

    var ms micro.Service
    options := []micro.Option{
        micro.Name(pkg.ServiceName),
        micro.Version(pkg.Version),
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

    if cfg.MicroRegistry == constant.RegistryKubernetes {
        ms = k8s.NewService(options...)
        logger.Info("Initialize k8s service")
    } else {
        ms = micro.NewService(options...)
        logger.Info("Initialize micro service")
    }

    ms.Init()

    err = currencies.RegisterCurrencyratesServiceHandler(ms.Server(), cs)
    if err != nil {
        logger.Fatal("Can`t register service in micro", zap.Error(err))
    }

    if err := ms.Run(); err != nil {
        logger.Fatal("Can`t run service", zap.Error(err))
    }
}
