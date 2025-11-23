package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/test-avito/internal/db"
	"github.com/test-avito/internal/models"
	"github.com/test-avito/internal/service"
)

type Server struct {
	svc *service.Service
}

func NewServer(svc *service.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Post("/team/add", s.handleTeamAdd)
	r.Get("/team/get", s.handleTeamGet)

	r.Post("/users/setIsActive", s.handleSetUserActive)
	r.Get("/users/getReview", s.handleUserReviews)

	r.Post("/pullRequest/create", s.handleCreatePR)
	r.Post("/pullRequest/merge", s.handleMergePR)
	r.Post("/pullRequest/reassign", s.handleReassign)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return r
}

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func (s *Server) handleTeamAdd(w http.ResponseWriter, r *http.Request) {
	var req models.Team
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, buildError("INVALID_BODY", err.Error()))
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, buildError("INVALID_INPUT", "team_name is required"))
		return
	}
	team, err := s.svc.CreateTeam(req)
	if err != nil {
		status, code := mapError(err)
		writeJSON(w, status, buildError(code, err.Error()))
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"team": team})
}

func (s *Server) handleTeamGet(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeJSON(w, http.StatusBadRequest, buildError("INVALID_INPUT", "team_name is required"))
		return
	}
	team, err := s.svc.GetTeam(teamName)
	if err != nil {
		status, code := mapError(err)
		writeJSON(w, status, buildError(code, err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, team)
}

type setActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

func (s *Server) handleSetUserActive(w http.ResponseWriter, r *http.Request) {
	var req setActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, buildError("INVALID_BODY", err.Error()))
		return
	}
	if req.UserID == "" {
		writeJSON(w, http.StatusBadRequest, buildError("INVALID_INPUT", "user_id is required"))
		return
	}
	user, err := s.svc.SetUserActive(req.UserID, req.IsActive)
	if err != nil {
		status, code := mapError(err)
		writeJSON(w, status, buildError(code, err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}

type createPRRequest struct {
	ID     string `json:"pull_request_id"`
	Name   string `json:"pull_request_name"`
	Author string `json:"author_id"`
}

func (s *Server) handleCreatePR(w http.ResponseWriter, r *http.Request) {
	var req createPRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, buildError("INVALID_BODY", err.Error()))
		return
	}
	if req.ID == "" || req.Name == "" || req.Author == "" {
		writeJSON(w, http.StatusBadRequest, buildError("INVALID_INPUT", "pull_request_id, pull_request_name and author_id are required"))
		return
	}
	pr, err := s.svc.CreatePullRequest(models.PullRequest{
		ID:       req.ID,
		Name:     req.Name,
		AuthorID: req.Author,
	})
	if err != nil {
		status, code := mapError(err)
		writeJSON(w, status, buildError(code, err.Error()))
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"pr": pr})
}

type mergePRRequest struct {
	ID string `json:"pull_request_id"`
}

func (s *Server) handleMergePR(w http.ResponseWriter, r *http.Request) {
	var req mergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, buildError("INVALID_BODY", err.Error()))
		return
	}
	if req.ID == "" {
		writeJSON(w, http.StatusBadRequest, buildError("INVALID_INPUT", "pull_request_id is required"))
		return
	}
	pr, err := s.svc.MergePullRequest(req.ID)
	if err != nil {
		status, code := mapError(err)
		writeJSON(w, status, buildError(code, err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"pr": pr})
}

type reassignRequest struct {
	PRID        string `json:"pull_request_id"`
	OldReviewer string `json:"old_user_id"`
}

func (s *Server) handleReassign(w http.ResponseWriter, r *http.Request) {
	var req reassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, buildError("INVALID_BODY", err.Error()))
		return
	}
	if req.PRID == "" || req.OldReviewer == "" {
		writeJSON(w, http.StatusBadRequest, buildError("INVALID_INPUT", "pull_request_id and old_user_id are required"))
		return
	}
	pr, replacedBy, err := s.svc.ReassignReviewer(req.PRID, req.OldReviewer)
	if err != nil {
		status, code := mapError(err)
		writeJSON(w, status, buildError(code, err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"pr":          pr,
		"replaced_by": replacedBy,
	})
}

func (s *Server) handleUserReviews(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeJSON(w, http.StatusBadRequest, buildError("INVALID_INPUT", "user_id is required"))
		return
	}
	prs, err := s.svc.ListUserReviews(userID)
	if err != nil {
		status, code := mapError(err)
		writeJSON(w, status, buildError(code, err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	})
}

func buildError(code, message string) errorResponse {
	var resp errorResponse
	resp.Error.Code = code
	resp.Error.Message = message
	return resp
}

func mapError(err error) (int, string) {
	switch err {
	case db.ErrTeamExists:
		return http.StatusBadRequest, "TEAM_EXISTS"
	case db.ErrTeamNotFound:
		return http.StatusNotFound, "NOT_FOUND"
	case db.ErrUserNotFound:
		return http.StatusNotFound, "NOT_FOUND"
	case db.ErrPRExists:
		return http.StatusConflict, "PR_EXISTS"
	case db.ErrPRNotFound:
		return http.StatusNotFound, "NOT_FOUND"
	case service.ErrPRMerged:
		return http.StatusConflict, "PR_MERGED"
	case service.ErrReviewerMissing:
		return http.StatusConflict, "NOT_ASSIGNED"
	case service.ErrNoCandidate:
		return http.StatusConflict, "NO_CANDIDATE"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR"
	}
}
