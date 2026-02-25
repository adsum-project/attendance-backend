package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/adsum-project/attendance-backend/internal/handlers"
)

type Handler func(w http.ResponseWriter, r *http.Request)
type Middleware func(Handler) Handler

type Route struct {
	httpMethod string
	pattern    string
	handler    Handler
	middleware []Middleware
}

type Router struct {
	mux              *http.ServeMux
	routes           []Route
	prefix           string
	globalMiddleware []Middleware
}

func NewRouter() *Router {
	return &Router{
		mux:              http.NewServeMux(),
		routes:           []Route{},
		prefix:           "",
		globalMiddleware: []Middleware{},
	}
}

func (r *Router) createMuxHandlers() {
	for _, route := range r.routes {
		handler := route.handler
		for _, m := range route.middleware {
			handler = m(handler)
		}

		pattern := route.pattern
		if route.httpMethod != "" {
			pattern = route.httpMethod + " " + route.pattern
		}
		r.mux.HandleFunc(pattern, handler)
	}
}

func (r *Router) StartServer(addr string) {

	r.createMuxHandlers()

	baseHandler := func(w http.ResponseWriter, req *http.Request) {
		if len(req.URL.Path) > 1 && strings.HasSuffix(req.URL.Path, "/") {
			req.URL.Path = strings.TrimSuffix(req.URL.Path, "/")
		}

		path := req.URL.Path
		method := req.Method

		if r.pathNotExists(path) {
			handlers.NotFound(w, req)
			return
		}

		if r.methodNotAllowed(path, method) {
			handlers.MethodNotAllowed(w, req)
			return
		}

		r.mux.ServeHTTP(w, req)
	}

	handler := baseHandler
	for _, middleware := range r.globalMiddleware {
		handler = middleware(handler)
	}

	fmt.Println("Server starting! Listening on", addr)
	http.ListenAndServe(addr, http.HandlerFunc(handler))
}

func (r *Router) createRoute(httpMethod string, pattern string, handler Handler) *Route {
	route := &Route{
		httpMethod: httpMethod,
		pattern:    r.prefix + pattern,
		handler:    handler,
		middleware: []Middleware{},
	}

	idx := len(r.routes)
	r.routes = append(r.routes, *route)
	return &r.routes[idx]
}

func (r *Router) Get(pattern string, handler Handler) *Route {
	return r.createRoute("GET", pattern, handler)
}

func (r *Router) Post(pattern string, handler Handler) *Route {
	return r.createRoute("POST", pattern, handler)
}

func (r *Router) Put(pattern string, handler Handler) *Route {
	return r.createRoute("PUT", pattern, handler)
}

func (r *Router) Delete(pattern string, handler Handler) *Route {
	return r.createRoute("DELETE", pattern, handler)
}

func (r *Router) Patch(pattern string, handler Handler) *Route {
	return r.createRoute("PATCH", pattern, handler)
}

func (rt *Route) Use(middleware Middleware) *Route {
	rt.middleware = append(rt.middleware, middleware)
	return rt
}

func (r *Router) Group(prefix string, methods func()) {
	prevPrefix := r.prefix
	r.prefix = prevPrefix + prefix
	methods()
	r.prefix = prevPrefix
}

func (r *Router) Use(middleware Middleware) {
	r.globalMiddleware = append(r.globalMiddleware, middleware)
}

func (r *Router) SetPrefix(prefix string) {
	r.prefix = prefix
}
