package analytics

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gasflow/api/internal/gateway"
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
	kpi, err := h.svc.GetKPIs(r.Context(), from, to)
	if err != nil {
		gateway.InternalError(w, err.Error())
		return
	}
	gateway.OK(w, kpi)
}

func (h *Handler) deliveries(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := DeliveryFilter{
		Status:   q.Get("status"),
		DriverID: q.Get("driver_id"),
	}
	if v := q.Get("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.From = &t
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			end := t.Add(24*time.Hour - time.Second)
			f.To = &end
		}
	}
	rows, total, err := h.svc.ListDeliveries(r.Context(), f)
	if err != nil {
		gateway.InternalError(w, err.Error())
		return
	}
	gateway.OK(w, map[string]any{"data": rows, "total": total})
}

func (h *Handler) driverPerformance(w http.ResponseWriter, r *http.Request) {
	from, to := parsePeriod(r)
	data, err := h.svc.DriverPerformance(r.Context(), from, to)
	if err != nil {
		gateway.InternalError(w, err.Error())
		return
	}
	gateway.OK(w, data)
}

func (h *Handler) topClients(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.TopClients(r.Context(), 10)
	if err != nil {
		gateway.InternalError(w, err.Error())
		return
	}
	gateway.OK(w, data)
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