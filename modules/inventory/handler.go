package inventory

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/deposits", h.listDeposits)
	r.Get("/deposits/{depositID}/items", h.listItems)
	r.Post("/deposits/{depositID}/receive", h.receive)
	r.Get("/low-stock", h.lowStock)
}

func (h *Handler) listDeposits(w http.ResponseWriter, r *http.Request) {
	deposits, err := h.svc.ListDeposits(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, deposits)
}

func (h *Handler) listItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListByDeposit(r.Context(), chi.URLParam(r, "depositID"))
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, items)
}

func (h *Handler) receive(w http.ResponseWriter, r *http.Request) {
	var in ReceiveInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		jsonError(w, "payload inválido", http.StatusBadRequest)
		return
	}
	in.DepositID = chi.URLParam(r, "depositID")
	if err := h.svc.Receive(r.Context(), in); err != nil {
		jsonError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	jsonOK(w, map[string]string{"status": "received"})
}

func (h *Handler) lowStock(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.LowStockItems(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, items)
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