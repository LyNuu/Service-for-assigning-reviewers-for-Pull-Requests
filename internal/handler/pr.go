package handler

import (
	"avitoMerchStore/internal/model"
	"encoding/json"
	"net/http"
	"strings"
)

type PrHandler struct {
	service prService
}

func NewPrHandler(service prService) *PrHandler {
	return &PrHandler{service: service}
}

type PRStatus string
type PullRequest struct {
	ID        string   `json:"pull_request_id"`
	Name      string   `json:"pull_request_name"`
	AuthorID  string   `json:"author_id"`
	Status    PRStatus `json:"status"`
	Reviewers []string `json:"assigned_reviewers"`
}

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

func (h *PrHandler) CreatePR(w http.ResponseWriter, r *http.Request) {

	var p PullRequest
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pr := &model.PullRequest{
		ID:        p.ID,
		Name:      p.Name,
		AuthorID:  p.AuthorID,
		Status:    model.StatusOpen,
		Reviewers: []string{},
	}
	res, err := h.service.CreatePR(r.Context(), pr)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Команда не найдена") {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Автор/команда не найдены"))
			return
		}
		if strings.Contains(err.Error(), "PR already exists") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorDetail{
					Code:    "PR_EXISTS",
					Message: "PR id already exists",
				},
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	p.Reviewers = res.Reviewers
	p.Status = PRStatus(res.Status)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pr": p,
	})
}

func (h *PrHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	var p struct {
		Id string `json:"pull_request_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.service.MergePR(r.Context(), p.Id)
	if err != nil {
		if strings.Contains(err.Error(), "Pull request not found") {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("PR не найден"))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var pr PullRequest
	pr.ID = res.ID
	pr.Name = res.Name
	pr.AuthorID = res.AuthorID
	pr.Status = PRStatus(res.Status)
	pr.Reviewers = res.Reviewers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("PR в состоянии MERGED"))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pr": pr,
	})
}

func (h *PrHandler) ReassignPR(w http.ResponseWriter, r *http.Request) {
	var pr struct {
		ID    string `json:"pull_request_id"`
		OldId string `json:"old_reviewer_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&pr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p := model.PullRequest{
		ID: pr.ID,
	}
	res, err := h.service.ReassignPR(r.Context(), &p, pr.OldId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("PR или пользователь не найден"))
			return
		}
		if strings.Contains(err.Error(), "already merged") {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("Нарушение доменных правил переназначения"))
			json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorDetail{
					Code:    "PR_MERGED",
					Message: "cannot reassign on merged PR",
				},
			})
			return
		}
		if strings.Contains(err.Error(), "reviewer is not assigned") {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("Нарушение доменных правил переназначения"))
			json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorDetail{
					Code:    "NOT_ASSIGNED",
					Message: "reviewer is not assigned to this PR",
				},
			})
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resPr := PullRequest{
		ID:        res.ID,
		Name:      res.Name,
		AuthorID:  res.AuthorID,
		Status:    PRStatus(res.Status),
		Reviewers: res.Reviewers,
	}
	if resPr.Reviewers == nil {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("Нарушение доменных правил переназначения"))
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: ErrorDetail{
				Code:    "NO_CANDIDATE",
				Message: "no active replacement candidate in team",
			},
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pr": resPr,
	})
}
