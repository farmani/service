package handlers

import (
	"net/http"
	"os"

	"github.com/farmani/service/foundation/web"

	"github.com/farmani/service/app/services/sales-api/handlers/v1/testgrp"
	"github.com/farmani/service/business/web/v1/middlewares"

	"go.uber.org/zap"
)

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Build    string
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
}

// APIMux constructs a http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig) *web.App {

	mux := web.NewApp(cfg.Shutdown, middlewares.Logger(cfg.Log))

	mux.Handle(http.MethodGet, "/test", testgrp.Test)

	return mux
}
