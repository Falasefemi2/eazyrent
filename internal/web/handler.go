package web

import (
	"encoding/json"
	"net/http"

	"github.com/femi/golang-easyrent/internal/auth"
	"github.com/femi/golang-easyrent/internal/listing"

	_ "github.com/femi/golang-easyrent/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Handler struct {
	AuthSvc  auth.Service
	Auth     Auth
	Listings listing.Service
}

func NewHandler(authSvc auth.Service, listings listing.Service) Handler {
	return Handler{
		AuthSvc:  authSvc,
		Auth:     Auth{Users: authSvc.Users, Tokens: authSvc.Tokens},
		Listings: listings,
	}
}

func (h Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /auth/signup", h.signUp)
	mux.HandleFunc("POST /auth/signin", h.signIn)
	mux.HandleFunc("POST /auth/refresh", h.refresh)
	mux.HandleFunc("POST /auth/signout", h.signOut)
	mux.HandleFunc("GET /auth/verify", h.verifyEmail)
	mux.HandleFunc("POST /auth/forgot-password", h.forgotPassword)
	mux.HandleFunc("POST /auth/reset-password", h.resetPassword)
	mux.HandleFunc("GET /me", h.Auth.RequireAuth(h.me))
	mux.HandleFunc("PUT /me/avatar", h.Auth.RequireAuth(h.updateAvatar))
	mux.HandleFunc("GET /listings", h.listListings)
	mux.HandleFunc("GET /listings/my", h.Auth.RequireAuth(h.myListings))
	mux.HandleFunc("GET /listings/{id}", h.getListing)
	mux.HandleFunc("POST /listings", h.Auth.RequireAuth(h.createListing))
	mux.HandleFunc("PATCH /listings/{id}", h.Auth.RequireAuth(h.updateListing))
	mux.HandleFunc("DELETE /listings/{id}", h.Auth.RequireAuth(h.deleteListing))
	mux.HandleFunc("PATCH /listings/{id}/status", h.Auth.RequireAuth(h.updateListingStatus))
	mux.HandleFunc("POST /listings/{id}/media", h.Auth.RequireAuth(h.addMedia))
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	return mux
}

func (h Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}
