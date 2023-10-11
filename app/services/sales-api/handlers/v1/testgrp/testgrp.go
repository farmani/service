package testgrp

import (
	"context"
	"github.com/farmani/service/foundation/web"
	"net/http"
)

// Test handles GET requests to test endpoint.
func Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	status := struct {
		Status string
	}{
		Status: "OK",
	}

	return web.Respond(ctx, w, status, http.StatusOK)
}
