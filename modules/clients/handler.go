package clients

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.getByID)
	r.Post("/{id}/block", h.block)
	r.Post("/{id}/activate", h.activate)
	r.Post("/{id}/addresses", h.addAddress)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	if limit == 0 {
		limit = 20
	}

	clients, total, err := h.svc.List(r.Context(), ListFilter{
		Status: q.Get("status"),
		Search: q.Get("search"),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]any{"data": clients, "total": total, "limit": limit, "offset": offset})
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in CreateClientInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		jsonError(w, "payload inválido", http.StatusBadRequest)
		return
	}

	c, err := h.svc.Create(r.Context(), in)
	if err != nil {
		log.Error().Err(err).Msg("clients: create")
		jsonError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	jsonOK(w, c)
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		jsonError(w, "cliente não encontrado", http.StatusNotFound)
		return
	}
	jsonOK(w, c)
}

func (h *Handler) block(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Block(r.Context(), id); err != nil {
		jsonError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	jsonOK(w, map[string]string{"status": "blocked"})
}

func (h *Handler) activate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Activate(r.Context(), id); err != nil {
		jsonError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	jsonOK(w, map[string]string{"status": "active"})
}

func (h *Handler) addAddress(w http.ResponseWriter, r *http.Request) {
	var in AddAddressInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		jsonError(w, "payload inválido", http.StatusBadRequest)
		return
	}
	in.ClientID = chi.URLParam(r, "id")

	addr, err := h.svc.AddAddress(r.Context(), in)
	if err != nil {
		jsonError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	jsonOK(w, addr)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func jsonOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}