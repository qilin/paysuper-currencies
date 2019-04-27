package main

import (
    "fmt"
    "github.com/InVisionApp/go-health"
    "github.com/InVisionApp/go-health/handlers"
    "github.com/globalsign/mgo"
    "github.com/micro/go-micro"
    "github.com/micro/go-plugins/wrapper/monitoring/prometheus"
    k8s "github.com/micro/kubernetes/go/micro"
    "github.com/paysuper/paysuper-currencies-rates/config"
    "github.com/paysuper/paysuper-currencies-rates/internal"
    "github.com/paysuper/paysuper-currencies-rates/pkg"
    currencyrates "github.com/paysuper/paysuper-currencies-rates/proto"
    "github.com/paysuper/paysuper-currencies-rates/utils"
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

    mongoUrl := utils.GetMongoUrl(cfg)
    session, err := mgo.Dial(mongoUrl)
    if err != nil {
        logger.Fatal(err.Error())
        return
    }
    defer session.Close()

    db := session.DB(cfg.MongoDatabase)

    cs, err := internal.NewService(cfg, db)
    if err != nil {
        logger.Fatal("Can`t create currency rates service", zap.Error(err))
    }

    cs.Init()
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
