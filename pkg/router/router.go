package router

import (
	"net/http"

	"github.com/adsum-project/attendance-backend/internal/handlers"
)

type Handler func(w http.ResponseWriter, r *http.Request)
type Middleware func(Handler) Handler

type Route struct {
	httpMethod string
	pattern string
	handler Handler
	middleware []Middleware
}

type Router struct {
	mux *http.ServeMux
	routes []Route
	prefix string
}

func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
		routes: []Route{},
		prefix: "",
	}
}

func (r *Router) createMuxHandlers() {
	for _, route := range r.routes {
		handler := route.handler
		for _, middleware := range route.middleware {
			handler = middleware(handler)
		}
		
		r.mux.HandleFunc(route.pattern, func(w http.ResponseWriter, req *http.Request) {
			if route.httpMethod != "" && route.httpMethod != req.Method {
				handlers.MethodNotAllowed(w, req)
				return
			}
			handler(w, req)
		})
	}
}

func (r *Router) StartServer(addr string) {
	r.createMuxHandlers()
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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
	})
	
	http.ListenAndServe(addr, handler)
}

func (r *Router) createRoute(httpMethod string, pattern string, handler Handler) *Route {
	route := &Route{
		httpMethod: httpMethod,
		pattern: r.prefix + pattern,
		handler: handler,
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

func (r *Router) SetPrefix(prefix string) {
	r.prefix = prefix
}