package router

import (
	"errors"
	"fmt"

	"github.com/adsum-project/attendance-backend/pkg/utils"
)

func (r *Router) pathNotExists(path string) bool {
	for _, route := range r.routes {
		if route.pattern == path {
			return false
		}
	}
	return true
}

func (r *Router) methodNotAllowed(path string, method string) bool {
	for _, route := range r.routes {
		if route.pattern == path {
			if route.httpMethod == "" || route.httpMethod == method {
				return false
			}
		}
	}
	return true
}

func (r *Router) DebugAllRoutes() error {
	if utils.GetEnvironment() == "production" {
		return errors.New("unable to debug router in production environment")
	}

	for _, route := range r.routes {
		httpMethod := route.httpMethod
		if route.httpMethod == "" {
			httpMethod = "ALL METHODS"
		}

		fmt.Println(httpMethod + " " + route.pattern)
	}
	return nil
}
