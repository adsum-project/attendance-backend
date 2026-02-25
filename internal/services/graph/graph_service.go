package graph

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2/clientcredentials"
)

const graphBaseURL = "https://graph.microsoft.com/v1.0"

type GraphService struct {
	httpClient *http.Client
}

func NewGraphService() (*GraphService, error) {
	clientID := os.Getenv("ENTRA_CLIENT_ID")
	if clientID == "" {
		return nil, fmt.Errorf("ENTRA_CLIENT_ID is required for Graph API")
	}
	clientSecret := os.Getenv("ENTRA_CLIENT_SECRET")
	if clientSecret == "" {
		return nil, fmt.Errorf("ENTRA_CLIENT_SECRET is required for Graph API")
	}
	tenantID := os.Getenv("ENTRA_TENANT_ID")
	if tenantID == "" {
		return nil, fmt.Errorf("ENTRA_TENANT_ID is required for Graph API")
	}

	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID),
		Scopes:       []string{"https://graph.microsoft.com/.default"},
	}

	ctx := context.Background()
	httpClient := config.Client(ctx)
	httpClient.Timeout = 30 * time.Second

	return &GraphService{
		httpClient: httpClient,
	}, nil
}
