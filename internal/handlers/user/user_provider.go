package userhandlers

import (
	"fmt"

	"github.com/adsum-project/attendance-backend/internal/services/graph"
)

type UserProvider struct {
	graph *graph.GraphService
}

func NewUserProvider(g *graph.GraphService) (*UserProvider, error) {
	if g == nil {
		return nil, fmt.Errorf("graph service is required")
	}

	return &UserProvider{
		graph: g,
	}, nil
}
