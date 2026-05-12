package analytics

import (
	"net/http"
	"time"


	"github.com/go-chi/chi/v5"
	"gasflow/internal/httputil"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Routes(r chi.Router) {
	r.Get("/kpis", h.kpis)
	r.Get("/deliveries", h.deliveries)
	r.Get("/driver-performance", h.driverPerformance)
	r.Get("/top-clients", h.topClients)
}

func (h *Handler) kpis(w http.ResponseWriter, r *http.Request) {
	from, to := parsePeriod(r)
	kpi, err := h.svc.GetKPIs(r.Context(), &from, &to)
	if err != nil {
		// Adicionado o status code 500
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.OK(w, kpi)
}

func (h *Handler) deliveries(w http.ResponseWriter, r *http.Request) {
	// ... (código anterior)
	f := DeliveryFilter{}
	rows, total, err := h.svc.ListDeliveries(r.Context(),  f)
	if err != nil {
		// Adicionado o status code 500
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.OK(w, map[string]any{"data": rows, "total": total})
}

func (h *Handler) driverPerformance(w http.ResponseWriter, r *http.Request) {
	from, to := parsePeriod(r)
	// Adicionado '&' para passar como ponteiro *time.Time
	data, err := h.svc.DriverPerformance(r.Context(), &from, &to)
	if err != nil {
		// Adicionado o status code 500
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.OK(w, data)
}

func (h *Handler) topClients(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.TopClients(r.Context(), 10)
	if err != nil {
		// httputil.Error(w, err.Error())
		return
	}
	httputil.OK(w, data)
}

func parsePeriod(r *http.Request) (time.Time, time.Time) {
	to := time.Now().UTC()
	from := to.AddDate(0, -1, 0)
	q := r.URL.Query()
	if v := q.Get("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			from = t
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			to = t.Add(24*time.Hour - time.Second)
		}
	}
	return from, to
}