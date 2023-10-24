// Package debug provides handler support for the debugging endpoints.
package debug

import (
	"expvar"
	"net/http"
	"net/http/pprof"

	"github.com/farmani/service/app/services/sales-api/handlers/v1/checkgrp"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Mux registers all the debug routes from the standard library into a new mux
// bypassing the use of the DefaultServerMux. Using the DefaultServerMux would
// be a security risk since a dependency could inject a handler into our service
// without us knowing it.
func StandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}

func Mux(build string, log *zap.SugaredLogger, db *sqlx.DB) http.Handler {
	mux := StandardLibraryMux()

	chgrp := checkgrp.Handlers{
		Build: build,
		Log:   log,
		DB:    db,
	}

	mux.HandleFunc("/v1/debug/readiness", chgrp.Readiness)
	mux.HandleFunc("/v1/debug/liveness", chgrp.Liveness)

	return mux
}
