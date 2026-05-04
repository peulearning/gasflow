package analytics

import "fmt"

// ─────────────────────────────────────────────────────────────
// GENERIC RESPONSES
// ─────────────────────────────────────────────────────────────

type ListMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type APIResponse struct {
	Data any `json:"data"`
	Meta any `json:"meta,omitempty"`
}

// ─────────────────────────────────────────────────────────────
// KPI RESPONSE
// ─────────────────────────────────────────────────────────────

type KPIResponse struct {
	Period     PeriodInfo         `json:"period"`
	Deliveries DeliveryKPIView    `json:"deliveries"`
	Inventory  InventoryKPIView   `json:"inventory"`
	Billing    BillingKPIView     `json:"billing"`
}

type DeliveryKPIView struct {
	Total       int     `json:"total"`
	Delivered   int     `json:"delivered"`
	Delayed     int     `json:"delayed"`
	Rescheduled int     `json:"rescheduled"`
	SLARate     string  `json:"sla_rate"`
	VolumeKG    string  `json:"volume_kg"`
}

type InventoryKPIView struct {
	TotalUnits     int `json:"total_units"`
	Reserved       int `json:"reserved"`
	Available      int `json:"available"`
	LowStockAlerts int `json:"low_stock_alerts"`
}

type BillingKPIView struct {
	RevenueCents       int64  `json:"revenue_cents"`
	RevenueFormatted   string `json:"revenue_formatted"`
	OverdueCount       int    `json:"overdue_count"`
	OverdueAmountCents int64  `json:"overdue_amount_cents"`
	OverdueFormatted   string `json:"overdue_formatted"`
}

func MaterializeKPIs(k KPISummary) APIResponse {
	return APIResponse{
		Data: KPIResponse{
			Period: k.Period,
			Deliveries: DeliveryKPIView{
				Total:       k.Deliveries.Total,
				Delivered:   k.Deliveries.Delivered,
				Delayed:     k.Deliveries.Delayed,
				Rescheduled: k.Deliveries.Rescheduled,
				SLARate:     fmt.Sprintf("%.2f%%", k.Deliveries.SLARate),
				VolumeKG:    fmt.Sprintf("%.2f kg", k.Deliveries.VolumeKG),
			},
			Inventory: InventoryKPIView{
				TotalUnits:     k.Inventory.TotalUnits,
				Reserved:       k.Inventory.Reserved,
				Available:      k.Inventory.Available,
				LowStockAlerts: k.Inventory.LowStockAlerts,
			},
			Billing: BillingKPIView{
				RevenueCents:       k.Billing.RevenueCents,
				RevenueFormatted:   centsToMoney(k.Billing.RevenueCents),
				OverdueCount:       k.Billing.OverdueCount,
				OverdueAmountCents: k.Billing.OverdueAmountCents,
				OverdueFormatted:   centsToMoney(k.Billing.OverdueAmountCents),
			},
		},
	}
}

// ─────────────────────────────────────────────────────────────
// DELIVERIES RESPONSE
// ─────────────────────────────────────────────────────────────

func MaterializeDeliveries(rows []DeliveryRow, total, limit, offset int) APIResponse {
	return APIResponse{
		Data: rows,
		Meta: ListMeta{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}
}

// ─────────────────────────────────────────────────────────────
// DRIVER PERFORMANCE RESPONSE
// ─────────────────────────────────────────────────────────────

type DriverPerformanceView struct {
	DriverID   string `json:"driver_id"`
	DriverName string `json:"driver_name"`
	Total      int    `json:"total"`
	Delivered  int    `json:"delivered"`
	Delayed    int    `json:"delayed"`
	SLARate    string `json:"sla_rate"`
}

func MaterializeDriverPerformance(rows []DriverPerf) APIResponse {
	var result []DriverPerformanceView

	for _, r := range rows {
		result = append(result, DriverPerformanceView{
			DriverID:   r.DriverID,
			DriverName: r.DriverName,
			Total:      r.Total,
			Delivered:  r.Delivered,
			Delayed:    r.Delayed,
			SLARate:    fmt.Sprintf("%.2f%%", r.SLARate),
		})
	}

	return APIResponse{
		Data: result,
	}
}

// ─────────────────────────────────────────────────────────────
// TOP CLIENTS RESPONSE
// ─────────────────────────────────────────────────────────────

type TopClientView struct {
	ClientID        string `json:"client_id"`
	ClientName      string `json:"client_name"`
	TotalOrders     int    `json:"total_orders"`
	TotalCents      int64  `json:"total_cents"`
	TotalFormatted  string `json:"total_formatted"`
}

func MaterializeTopClients(rows []TopClient) APIResponse {
	var result []TopClientView

	for _, r := range rows {
		result = append(result, TopClientView{
			ClientID:       r.ClientID,
			ClientName:     r.ClientName,
			TotalOrders:    r.TotalOrders,
			TotalCents:     r.TotalCents,
			TotalFormatted: centsToMoney(r.TotalCents),
		})
	}

	return APIResponse{
		Data: result,
	}
}

// ─────────────────────────────────────────────────────────────
// MONEY FORMATTER
// ─────────────────────────────────────────────────────────────

func centsToMoney(v int64) string {
	return fmt.Sprintf("R$ %.2f", float64(v)/100)
}