package apm

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/naughtygopher/errors"
	"github.com/naughtygopher/goapp/internal/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
)

func prometheusExporter() (*prometheus.Exporter, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, errors.Wrap(err, "promexporter.New")
	}

	return exporter, nil
}

func prometheusScraper(opts *Options) {
	mux := http.NewServeMux()
	mux.Handle("/-/metrics", promhttp.Handler())
	server := &http.Server{
		Handler:           mux,
		Addr:              fmt.Sprintf(":%d", opts.PrometheusScrapePort),
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info(context.Background(), fmt.Sprintf("[otel/http] starting prometheus metrics on :%d/-/metrics", opts.PrometheusScrapePort))
	err := server.ListenAndServe()
	if err != nil {
		logger.Error(context.Background(), fmt.Sprintf("[otel/http] failed to start prometheus metrics on :%d/-/metrics ; %+v", opts.PrometheusScrapePort, err))
	}
}
