package verificationhandlers

import (
	"io"
	"net/http"

	"github.com/adsum-project/attendance-backend/pkg/utils/response"
)

func (p *VerificationProvider) CreateEmbedding(w http.ResponseWriter, r *http.Request) {
	p.proxyRequest(w, r, "/embeddings")
}

func (p *VerificationProvider) VerifyEmbedding(w http.ResponseWriter, r *http.Request) {
	p.proxyRequest(w, r, "/embeddings/verify")
}

func (p *VerificationProvider) proxyRequest(w http.ResponseWriter, r *http.Request, upstreamPath string) {
	if p == nil || p.client == nil {
		response.InternalServerError(w, "Verification service not configured")
		return
	}

	url := p.embeddingsApiURL + upstreamPath
	if r.URL.RawQuery != "" {
		url += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, url, r.Body)
	if err != nil {
		response.InternalServerError(w, "Failed to create verification request")
		return
	}

	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		response.InternalServerError(w, "Failed to reach verification service")
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
