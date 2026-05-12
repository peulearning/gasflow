package analytics

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidPeriod = errors.New("analytics: período inválido")

type Service struct {
	repo *Repository
}

func (s *Service) TopClients(context context.Context, i int) (any, any) {
	panic("unimplemented")
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ─────────────────────────────────────────────────────────────
// KPI SUMMARY
// ─────────────────────────────────────────────────────────────

func (s *Service) GetKPIs(ctx context.Context, from, to *time.Time) (KPISummary, error) {
	start, end, err := normalizePeriod(from, to)
	if err != nil {
		return KPISummary{}, err
	}

	return s.repo.GetKPIs(ctx, start, end)
}

// ─────────────────────────────────────────────────────────────
// LIST DELIVERIES
// ─────────────────────────────────────────────────────────────

func (s *Service) ListDeliveries(ctx context.Context, f DeliveryFilter) ([]DeliveryRow, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	return s.repo.ListDeliveries(ctx, f)
}

// ─────────────────────────────────────────────────────────────
// DRIVER PERFORMANCE
// ─────────────────────────────────────────────────────────────

func (s *Service) DriverPerformance(ctx context.Context, from, to *time.Time) ([]DriverPerf, error) {
	start, end, err := normalizePeriod(from, to)
	if err != nil {
		return nil, err
	}

	return s.repo.DriverPerformance(ctx, start, end)
}

// ─────────────────────────────────────────────────────────────
// TOP CLIENTS
// ─────────────────────────────────────────────────────────────

func (s *Service) TopClientsByVolume(ctx context.Context, limit int) ([]TopClient, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.TopClientsByVolume(ctx, limit)
}

// ─────────────────────────────────────────────────────────────
// AUX
// ─────────────────────────────────────────────────────────────

func normalizePeriod(from, to *time.Time) (time.Time, time.Time, error) {
	now := time.Now()

	if from == nil && to == nil {
		start := now.AddDate(0, 0, -30)
		return start, now, nil
	}

	if from == nil || to == nil {
		return time.Time{}, time.Time{}, ErrInvalidPeriod
	}

	if from.After(*to) {
		return time.Time{}, time.Time{}, ErrInvalidPeriod
	}

	return *from, *to, nil
}
