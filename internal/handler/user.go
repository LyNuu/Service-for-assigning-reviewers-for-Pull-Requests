package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

type UserHandler struct {
	service userService
}

func NewUserHandler(service userService) *UserHandler {
	return &UserHandler{service: service}
}

type UserResponse struct {
	ID       string `json:"user_id"`
	Name     string `json:"username"`
	TeamName string `json:"team_name"`
	Status   bool   `json:"is_active"`
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string `json:"user_id"`
		Status bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := h.service.SetIsActive(r.Context(), req.ID, req.Status)
	if err != nil {
		if strings.Contains(err.Error(), "Пользователь не найден") {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	u := &UserResponse{
		ID:       user.ID,
		Name:     user.Name,
		TeamName: user.TeamName,
		Status:   user.Status,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"user": u})
}

func (h *UserHandler) GetPrById(w http.ResponseWriter, r *http.Request) {
	uID := r.URL.Query().Get("user_id")
	pr, err := h.service.GetPrById(r.Context(), uID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	type res struct {
		ID       string   `json:"pull_request_id"`
		Name     string   `json:"pull_request_name"`
		AuthorId string   `json:"author_id"`
		Status   PRStatus `json:"status"`
	}
	var sliceRes []res
	for _, v := range *pr {
		sliceRes = append(sliceRes, res{
			ID:       v.ID,
			Name:     v.Name,
			AuthorId: v.AuthorID,
			Status:   PRStatus(v.Status),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"pr": sliceRes})
}
