package middlewares

import (
	"context"
	"net/http"

	"github.com/farmani/service/business/web/auth"

	"github.com/farmani/service/business/sys/validate"
	v1 "github.com/farmani/service/business/web/v1"
	"github.com/farmani/service/foundation/web"
	"go.uber.org/zap"
)

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
func Errors(log *zap.SugaredLogger) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if err := handler(ctx, w, r); err != nil {
				log.Errorw("error", "trace_id", web.GetTraceID(ctx), "message", err)

				var er v1.ErrorResponse
				var status int

				switch {
				case validate.IsFieldErrors(err):
					fieldErrors := validate.GetFieldErrors(err)
					er = v1.ErrorResponse{
						Error:  "data validation error",
						Fields: fieldErrors.Fields(),
					}
					status = http.StatusBadRequest

				case v1.IsRequestError(err):
					reqErr := v1.GetRequestError(err)
					er = v1.ErrorResponse{
						Error: reqErr.Error(),
					}
					status = reqErr.Status

				case auth.IsAuthError(err):
					er = v1.ErrorResponse{
						Error: http.StatusText(http.StatusUnauthorized),
					}
					status = http.StatusUnauthorized

				default:
					er = v1.ErrorResponse{
						Error: http.StatusText(http.StatusInternalServerError),
					}
					status = http.StatusInternalServerError
				}

				// back propagate the error to the Handle method
				// we have to validate error there to be sure its because of shutdown or some other error
				// cause response could not be written to the client
				if err := web.Respond(ctx, w, er, status); err != nil {
					return err
				}

				// If we receive the shutdown err we need to return it
				// back to the base handler to shut down the service.
				if web.IsShutdown(err) {
					return err
				}
			}

			return nil
		}

		return h
	}

	return m
}
