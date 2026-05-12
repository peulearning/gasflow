package shared

import (
	"encoding/json"
	"time"
)

type AuditEntry struct {
	ID         string          `json:"id"`
	Entity     string          `json:"entity"`
	EntityID   string          `json:"entity_id"`
	Action     string          `json:"action"`
	UserID     string          `json:"user_id"`
	Payload    json.RawMessage `json:"payload"`
	OccurredAt time.Time       `json:"occurred_at"`
}

func NewAuditEntry(
	entity,
	entityID,
	action,
	userID string,
	payload interface{},
) (AuditEntry, error) {

	raw, err := json.Marshal(payload)
	if err != nil {
		return AuditEntry{}, err
	}

	return AuditEntry{
		Entity:     entity,
		EntityID:   entityID,
		Action:     action,
		UserID:     userID,
		Payload:    raw,
		OccurredAt: time.Now().UTC(),
	}, nil
}