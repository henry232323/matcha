package router

import (
	"net/http"

	"github.com/cloudretic/matcha/pkg/middleware"
	"github.com/cloudretic/matcha/pkg/rctx"
	"github.com/cloudretic/matcha/pkg/route"
	"github.com/cloudretic/matcha/pkg/tree"
)

type defaultRouter struct {
	mws      []middleware.Middleware
	routes   map[string]map[int]route.Route
	rtree    *tree.RouteTree
	handlers map[string]http.Handler
	notfound http.Handler
}

func Default() *defaultRouter {
	return &defaultRouter{
		mws:      make([]middleware.Middleware, 0),
		routes:   make(map[string]map[int]route.Route),
		rtree:    tree.New(),
		handlers: make(map[string]http.Handler),
		notfound: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) }),
	}
}

func (rt *defaultRouter) Attach(mw middleware.Middleware) {
	rt.mws = append(rt.mws, mw)
}

func (rt *defaultRouter) AddRoute(r route.Route, h http.Handler) {
	id := rt.rtree.Add(r)
	if rt.routes[r.Method()] == nil {
		rt.routes[r.Method()] = make(map[int]route.Route)
	}
	rt.routes[r.Method()][id] = r
	rt.handlers[r.Hash()] = h
}

func (rt *defaultRouter) AddNotFound(h http.Handler) {
	rt.notfound = h
}

// Implements http.Handler.
//
// Serve request using the registered middleware, routes, and handlers.
// Tree Router organizes routes by their 'prefixes' (first path elements) and serves based on the first
// path element of the request. Since wildcard and regex parts do not statically evaluate, they are stored as "*".
func (rt *defaultRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	req = executeMiddleware(rt.mws, w, req)
	if req == nil {
		return
	}
	req = rctx.PrepareRequestContext(req, rctx.DefaultMaxParams)
	leaf_id := rt.rtree.Match(req)
	if leaf_id != tree.NO_LEAF_ID {
		r := rt.routes[req.Method][leaf_id]
		reqWithCtx := r.MatchAndUpdateContext(req)
		reqWithCtx = executeMiddleware(r.Middleware(), w, reqWithCtx)
		if reqWithCtx == nil {
			return
		}
		rt.handlers[r.Hash()].ServeHTTP(w, reqWithCtx)
		return
	}
	rt.notfound.ServeHTTP(w, req)
	return
}

func executeMiddleware(mw []middleware.Middleware, w http.ResponseWriter, req *http.Request) *http.Request {
	for _, m := range mw {
		if req = m(w, req); req == nil {
			return nil
		}
	}
	return req
}
