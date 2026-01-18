package main

import (
	"context"
	"expvar"
	"net/http"
	"net/http/pprof"
	"os"
)

var (
	service = "projects-api"
	version = "0.0.1"
	environ = "development"
)

func main() {

	ctx := context.Background()
	if err := run(ctx); err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())
	return nil
}
