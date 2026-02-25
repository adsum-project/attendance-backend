package router

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/adsum-project/attendance-backend/pkg/utils"
)

func PathParam(r *http.Request, key string) string {
	if r == nil {
		return ""
	}
	return r.PathValue(key)
}

func splitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return []string{}
	}
	return strings.Split(trimmed, "/")
}

func isPathParamToken(token string) bool {
	if strings.HasPrefix(token, "{") && strings.HasSuffix(token, "}") && len(token) > 2 {
		return true
	}
	if strings.HasPrefix(token, ":") && len(token) > 1 {
		return true
	}
	return false
}

func matchPathPattern(pattern string, path string) bool {
	patternParts := splitPath(pattern)
	pathParts := splitPath(path)

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i := range patternParts {
		patternPart := patternParts[i]
		pathPart := pathParts[i]

		if isPathParamToken(patternPart) {
			continue
		}

		if patternPart != pathPart {
			return false
		}
	}

	return true
}

func (r *Router) pathNotExists(path string) bool {
	for _, route := range r.routes {
		if matchPathPattern(route.pattern, path) {
			return false
		}
	}
	return true
}

func (r *Router) methodNotAllowed(path string, method string) bool {
	for _, route := range r.routes {
		if matchPathPattern(route.pattern, path) {
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
