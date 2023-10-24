package testgrp

import (
	"context"
	"errors"
	v1 "github.com/farmani/service/business/web/v1"
	"github.com/farmani/service/foundation/web"
	"math/rand"
	"net/http"
)

// Test handles GET requests to test endpoint.
func Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if n := rand.Intn(100); n%2 == 0 {
		return v1.NewRequestError(errors.New("trusted error"), http.StatusBadRequest)
	}

	status := struct {
		Status string
	}{
		Status: "OK",
	}

	return web.Respond(ctx, w, status, http.StatusOK)
}
