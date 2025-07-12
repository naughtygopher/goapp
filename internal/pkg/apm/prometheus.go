package apm

import (
	"fmt"
	"net/http"
	"time"

	"github.com/naughtygopher/errors"
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
		Addr:              fmt.Sprintf("%d", opts.PrometheusScrapePort),
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Printf("[otel/http] starting prometheus metrics on :%d/-/metrics", opts.PrometheusScrapePort)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("[otel/http] failed to start prometheus metrics on :%d/-/metrics ; %+v", opts.PrometheusScrapePort, err)
		panic(err)
	}
}
