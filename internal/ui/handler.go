package ui

import (
	"embed"
	"errors"
	"html/template"
	"net/http"

	"github.com/google/uuid"
	"github.com/shophub-project-2026/shophub/internal/auth"
	"github.com/shophub-project-2026/shophub/internal/server/middleware"
	"github.com/shophub-project-2026/shophub/internal/shops"
)

//go:embed templates
var templatesFS embed.FS

type Handler struct {
	tmpl      *template.Template
	authSvc   *auth.Service
	shopsRepo shops.Repository
}

func NewHandler(authSvc *auth.Service, shopsRepo shops.Repository) *Handler {
	tmpl := template.Must(template.ParseFS(templatesFS, "templates/*.html"))
	return &Handler{tmpl: tmpl, authSvc: authSvc, shopsRepo: shopsRepo}
}

func (h *Handler) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	h.render(w, "login.html", map[string]any{})
}

func (h *Handler) LoginPost(w http.ResponseWriter, r *http.Request) {
	token, err := h.authSvc.Login(r.Context(), auth.LoginInput{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	})
	if err != nil {
		h.render(w, "login.html", map[string]any{"Error": "Invalid email or password"})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	h.render(w, "register.html", map[string]any{})
}

func (h *Handler) RegisterPost(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authSvc.Register(r.Context(), auth.RegisterInput{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}); err != nil {
		h.render(w, "register.html", map[string]any{"Error": "Registration failed. Email may already be in use."})
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) ShopEdit(w http.ResponseWriter, r *http.Request) {
	userID, _ := uuid.Parse(middleware.UserIDFromCtx(r.Context()))
	name := r.PathValue("name")

	list, err := h.shopsRepo.List(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to load shop", http.StatusInternalServerError)
		return
	}
	var found *shops.ShopView
	for i := range list {
		if list[i].Name == name {
			found = &list[i]
			break
		}
	}
	if found == nil {
		http.NotFound(w, r)
		return
	}
	h.render(w, "shop_edit.html", map[string]any{
		"Email": middleware.EmailFromCtx(r.Context()),
		"Shop":  *found,
	})
}

func (h *Handler) ShopEditPost(w http.ResponseWriter, r *http.Request) {
	userID, _ := uuid.Parse(middleware.UserIDFromCtx(r.Context()))
	name := r.PathValue("name")

	availability := r.FormValue("availability")
	walletAddress := r.FormValue("walletAddress")

	if _, err := h.shopsRepo.Update(r.Context(), userID, name, shops.UpdateInput{
		Availability:  &availability,
		WalletAddress: &walletAddress,
	}); err != nil {
		h.render(w, "shop_edit.html", map[string]any{
			"Email": middleware.EmailFromCtx(r.Context()),
			"Shop":  shops.ShopView{Name: name, Availability: availability, WalletAddress: walletAddress},
			"Error": "Failed to update shop: " + err.Error(),
		})
		return
	}
	http.Redirect(w, r, "/shops/"+name, http.StatusSeeOther)
}

func (h *Handler) ShopDetail(w http.ResponseWriter, r *http.Request) {
	userID, _ := uuid.Parse(middleware.UserIDFromCtx(r.Context()))
	name := r.PathValue("name")

	list, err := h.shopsRepo.List(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to load shop", http.StatusInternalServerError)
		return
	}
	var found *shops.ShopView
	for i := range list {
		if list[i].Name == name {
			found = &list[i]
			break
		}
	}
	if found == nil {
		http.NotFound(w, r)
		return
	}
	h.render(w, "shop_detail.html", map[string]any{
		"Email": middleware.EmailFromCtx(r.Context()),
		"Shop":  *found,
	})
}

func (h *Handler) ShopNew(w http.ResponseWriter, r *http.Request) {
	h.render(w, "shop_new.html", map[string]any{
		"Email": middleware.EmailFromCtx(r.Context()),
	})
}

func (h *Handler) ShopNewPost(w http.ResponseWriter, r *http.Request) {
	userID, _ := uuid.Parse(middleware.UserIDFromCtx(r.Context()))
	if _, err := h.shopsRepo.Create(r.Context(), userID, shops.CreateInput{
		Name:          r.FormValue("name"),
		Availability:  r.FormValue("availability"),
		WalletAddress: r.FormValue("walletAddress"),
		Database:      r.FormValue("database"),
	}); err != nil {
		msg := "Failed to create shop: " + err.Error()
		if errors.Is(err, shops.ErrNameTaken) {
			msg = "Shop name is already taken. Choose a different name."
		}
		h.render(w, "shop_new.html", map[string]any{
			"Email": middleware.EmailFromCtx(r.Context()),
			"Error": msg,
		})
		return
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	userID, _ := uuid.Parse(middleware.UserIDFromCtx(r.Context()))
	list, err := h.shopsRepo.List(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to load shops", http.StatusInternalServerError)
		return
	}
	h.render(w, "dashboard.html", map[string]any{
		"Email": middleware.EmailFromCtx(r.Context()),
		"Shops": list,
	})
}

func (h *Handler) Root(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if _, err := h.authSvc.ParseToken(cookie.Value); err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handler) ShopDelete(w http.ResponseWriter, r *http.Request) {
	userID, _ := uuid.Parse(middleware.UserIDFromCtx(r.Context()))
	name := r.PathValue("name")
	if err := h.shopsRepo.Delete(r.Context(), userID, name); err != nil {
		http.Error(w, "failed to delete shop", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
