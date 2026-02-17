package verificationhandlers

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/adsum-project/attendance-backend/pkg/utils"
	"github.com/adsum-project/attendance-backend/pkg/utils/response"
)

func (p *VerificationProvider) CreateEmbedding(w http.ResponseWriter, r *http.Request) {
	p.handleEmbeddingRequest(w, r)
}

func (p *VerificationProvider) GetEmbedding(w http.ResponseWriter, r *http.Request) {
	p.handleEmbeddingRequest(w, r)
}

func (p *VerificationProvider) UpdateEmbedding(w http.ResponseWriter, r *http.Request) {
	p.handleEmbeddingRequest(w, r)
}

func (p *VerificationProvider) DeleteEmbedding(w http.ResponseWriter, r *http.Request) {
	p.handleEmbeddingRequest(w, r)
}

func (p *VerificationProvider) VerifyEmbedding(w http.ResponseWriter, r *http.Request) {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	res, err := utils.RequestUpstream(
		r.Context(),
		p.client,
		r.Method,
		p.embeddingsApiURL,
		"/embeddings/verify",
		r.URL.Query(),
		headers,
		r.Body,
	)
	if err != nil {
		response.BadGateway(w, "Failed to reach verification service")
		return
	}

	var resBody interface{}
	if err := json.Unmarshal(res.Body, &resBody); err != nil {
		response.BadGateway(w, "Invalid response from verification service")
		return
	}

	response.JsonResponse(w, res.StatusCode, resBody)
}

func (p *VerificationProvider) handleEmbeddingRequest(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("userID").(string)

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	upstreamPath := "/embeddings/" + url.PathEscape(userID)

	res, err := utils.RequestUpstream(
		r.Context(),
		p.client,
		r.Method,
		p.embeddingsApiURL,
		upstreamPath,
		r.URL.Query(),
		headers,
		r.Body,
	)

	if err != nil {
		response.BadGateway(w, "Failed to reach verification service")
		return
	}

	if len(res.Body) == 0 {
		w.WriteHeader(res.StatusCode)
		return
	}

	var resBody interface{}
	if err := json.Unmarshal(res.Body, &resBody); err != nil {
		response.BadGateway(w, "Invalid response from verification service")
		return
	}

	response.JsonResponse(w, res.StatusCode, resBody)
}
