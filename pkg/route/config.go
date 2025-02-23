package route

import (
	"github.com/cloudretic/matcha/pkg/cors"
	"github.com/cloudretic/matcha/pkg/middleware"
)

// RouteConfigFuncs can be applied to a Route at creation.
type ConfigFunc func(Route) error

// Attaches middleware to the route that sets CORS headers on matched requests only.
func CORSHeaders(aco *cors.AccessControlOptions) ConfigFunc {
	return func(r Route) error {
		r.Attach(cors.CORSMiddleware(aco))
		return nil
	}
}

// Attaches middleware to the route.
func WithMiddleware(mw middleware.Middleware) ConfigFunc {
	return func(r Route) error {
		r.Attach(mw)
		return nil
	}
}
