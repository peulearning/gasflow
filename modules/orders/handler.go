package orders

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"gasflow/internal/infra/auth"
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
	r.Patch("/{id}/status", h.transition)
	r.Get("/{id}/history", h.getHistory)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := ListFilter{
		Status:   q.Get("status"),
		ClientID: q.Get("client_id"),
		DriverID: q.Get("driver_id"),
	}
	if v := q.Get("from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err == nil {
			f.From = &t
		}
	}
	if v := q.Get("to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err == nil {
			t = t.Add(24*time.Hour - time.Second)
			f.To = &t
		}
	}
	f.Limit, _ = strconv.Atoi(q.Get("limit"))
	f.Offset, _ = strconv.Atoi(q.Get("offset"))
	if f.Limit == 0 {
		f.Limit = 20
	}

	list, total, err := h.svc.List(r.Context(), f)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]any{"data": list, "total": total})
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in CreateOrderInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		jsonError(w, "payload inválido", http.StatusBadRequest)
		return
	}

	o, err := h.svc.Create(r.Context(), in)
	if err != nil {
		jsonError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, o)
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) {
	o, err := h.svc.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "pedido não encontrado", http.StatusNotFound)
		return
	}
	jsonOK(w, o)
}

func (h *Handler) transition(w http.ResponseWriter, r *http.Request) {
	var in TransitionInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		jsonError(w, "payload inválido", http.StatusBadRequest)
		return
	}
	in.OrderID = chi.URLParam(r, "id")

	// Pega o user do contexto de auth.
	if claims := auth.ClaimsFromCtx(r.Context()); claims != nil {
		in.ChangedBy = claims.UserID
	}

	o, err := h.svc.Transition(r.Context(), in)
	if err != nil {
		jsonError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	jsonOK(w, o)
}

func (h *Handler) getHistory(w http.ResponseWriter, r *http.Request) {
	history, err := h.svc.GetHistory(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, history)
}

func jsonOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}