package handler

import (
	"avitoMerchStore/internal/model"
	"encoding/json"
	"net/http"
	"strings"
)

type TeamHandler struct {
	service teamService
}

func NewTeamHandler(service teamService) *TeamHandler {
	return &TeamHandler{service: service}
}

type reqUser struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	Status   bool   `json:"is_active"`
}
type reqTeam struct {
	Name  string    `json:"team_name"`
	Users []reqUser `json:"members"`
}

func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	var req reqTeam
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	team := &model.Team{
		Team:  req.Name,
		Users: make([]model.User, len(req.Users)),
	}

	for i, user := range req.Users {
		team.Users[i] = model.User{
			ID:     user.ID,
			Name:   user.Username,
			Status: user.Status,
		}
	}

	_, err := h.service.AddTeam(r.Context(), team)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorDetail{
					Code:    "TEAM_EXISTS",
					Message: "team_name already exists",
				},
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"team": req})
	return
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	team := r.URL.Query().Get("team_name")
	if team == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	t, err := h.service.GetTeam(r.Context(), team)
	if err != nil {
		if strings.Contains(err.Error(), "Команда не найдена") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorDetail{
					Code:    "NOT_FOUND",
					Message: "team_name not found",
				},
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := reqTeam{
		Name:  t.Team,
		Users: make([]reqUser, len(t.Users)),
	}
	for i, user := range t.Users {
		resp.Users[i] = reqUser{
			ID:       user.ID,
			Username: user.Name,
			Status:   user.Status,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
