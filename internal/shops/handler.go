package shops

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/shophub-project-2026/shophub/internal/server/middleware"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	shops, err := h.repo.List(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to list shops", http.StatusInternalServerError)
		return
	}
	if shops == nil {
		shops = []ShopView{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(shops)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	var in CreateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if in.Name == "" || in.WalletAddress == "" {
		http.Error(w, "name and walletAddress are required", http.StatusBadRequest)
		return
	}

	shop, err := h.repo.Create(r.Context(), userID, in)
	if err != nil {
		if errors.Is(err, ErrNameTaken) {
			http.Error(w, "shop name already taken", http.StatusConflict)
			return
		}
		http.Error(w, "failed to create shop", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(shop)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	name := strings.TrimPrefix(r.PathValue("name"), "/")
	if name == "" {
		http.Error(w, "shop name is required", http.StatusBadRequest)
		return
	}

	var in UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	shop, err := h.repo.Update(r.Context(), userID, name, in)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "shop not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update shop", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(shop)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	name := strings.TrimPrefix(r.PathValue("name"), "/")
	if name == "" {
		http.Error(w, "shop name is required", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(r.Context(), userID, name); err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "shop not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete shop", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	raw := middleware.UserIDFromCtx(r.Context())
	id, err := uuid.Parse(raw)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return uuid.UUID{}, false
	}
	return id, true
}
