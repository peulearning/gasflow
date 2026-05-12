package billing

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)


type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/", h.list)
	r.Get("/overdue", h.overdue)
	r.Get("/{id}", h.getByID)
	r.Post("/{id}/pay", h.markPaid)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// Tratamento básico para paginação
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 10 // valor padrão
	}
	offset, _ := strconv.Atoi(q.Get("offset"))

	charges, total, err := h.svc.List(r.Context(), ListFilter{
		ClientID: q.Get("client_id"),
		Status:   q.Get("status"),
		Limit:    limit,
		Offset:   offset,
	})

	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]any{"data": charges, "total": total})
}

func (h *Handler) overdue(w http.ResponseWriter, r *http.Request) {
	// Reutilizando o Service com filtro fixo
	charges, total, err := h.svc.List(r.Context(), ListFilter{Status: "overdue", Limit: 100})
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]any{"data": charges, "total": total})
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		// Dica: verifique se o erro é de "não encontrado" ou erro de banco
		jsonError(w, "cobrança não encontrada", http.StatusNotFound)
		return
	}
	jsonOK(w, c)
}

func (h *Handler) markPaid(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.MarkPaid(r.Context(), id); err != nil {
		jsonError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	jsonOK(w, map[string]string{"status": "paid"})
}

// Funções auxiliares mantidas, com pequena melhoria de status
func jsonOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Garante o 200 OK
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}