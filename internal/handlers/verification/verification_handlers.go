package verificationhandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	verificationrepo "github.com/adsum-project/attendance-backend/internal/repositories/verification"
	"github.com/adsum-project/attendance-backend/internal/services/timetable"
	"github.com/adsum-project/attendance-backend/internal/services/verification"
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
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}
	if req.ImageBase64 == "" {
		response.BadRequest(w, "imageBase64 is required", nil)
		return
	}
	if req.ClassID == "" {
		response.BadRequest(w, "classId is required for attendance sign-in", nil)
		return
	}

	result, err := p.verificationService.VerifyEmbedding(r.Context(), req.ImageBase64, req.ClassID)
	if err != nil {
		if errors.Is(err, verification.ErrNoMatch) {
			response.NotFound(w, "No match")
			return
		}
		if errors.Is(err, timetable.ErrStudentNotEnrolled) {
			response.Forbidden(w, "You are not enrolled in this class")
			return
		}
		if errors.Is(err, timetable.ErrClassNotRunning) {
			response.Forbidden(w, "This class is not currently running")
			return
		}
		if errors.Is(err, verificationrepo.ErrAlreadySignedIn) {
			response.Forbidden(w, "You have already signed in to this class")
			return
		}
		response.InternalServerError(w, "Failed to verify")
		return
	}

	response.JsonResponse(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    result,
	})
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

func (p *VerificationProvider) QRStream(w http.ResponseWriter, r *http.Request) {
	classID := strings.TrimSpace(r.URL.Query().Get("classId"))
	if classID == "" {
		response.BadRequest(w, "classId is required", nil)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, canFlush := w.(http.Flusher)
	if canFlush {
		fmt.Fprint(w, ": connected\n\n")
		flusher.Flush()
	}

	ch, unsub := p.verificationService.QRTokenStream(classID)
	defer unsub()

	sendToken := func(t string) {
		fmt.Fprintf(w, "data: %s\n\n", t)
		if canFlush {
			flusher.Flush()
		}
	}

	token, err := p.verificationService.IssueQRToken(r.Context(), classID)
	if err != nil {
		response.InternalServerError(w, "Failed to create token")
		return
	}
	sendToken(token)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case t, ok := <-ch:
			if !ok {
				return
			}
			sendToken(t)
		case <-ticker.C:
			if _, err := p.verificationService.IssueQRToken(r.Context(), classID); err != nil {
				return
			}
		}
	}
}

func (p *VerificationProvider) QRVerify(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		response.BadRequest(w, "token is required", nil)
		return
	}

	userID, _ := r.Context().Value("userID").(string)
	if userID == "" {
		response.Unauthorized(w, "Sign in required")
		return
	}

	err := p.verificationService.SignInWithQRToken(r.Context(), userID, token)
	frontendURL := strings.TrimSuffix(strings.TrimSpace(os.Getenv("FRONTEND_URL")), "/")
	if frontendURL == "" {
		frontendURL = "/"
	}

	if err != nil {
		var status, message string
		switch {
		case errors.Is(err, verificationrepo.ErrTokenInvalid):
			status = "error"
			message = "Invalid or expired token"
		case errors.Is(err, timetable.ErrStudentNotEnrolled):
			status = "error"
			message = "You are not enrolled in this class"
		case errors.Is(err, timetable.ErrClassNotRunning):
			status = "error"
			message = "This class is not currently running"
		case errors.Is(err, verificationrepo.ErrAlreadySignedIn):
			status = "error"
			message = "You have already signed in to this class"
		default:
			status = "error"
			message = "Sign-in failed"
		}
		redirectURL := fmt.Sprintf("%s/attendance/qr-landing-page?status=%s&message=%s",
			frontendURL, url.QueryEscape(status), url.QueryEscape(message))
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	redirectURL := fmt.Sprintf("%s/attendance/qr-landing-page?status=success", frontendURL)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
