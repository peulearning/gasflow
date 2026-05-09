package order

import "errors"

type Status string

const (
	StatusReceived Status = "RECEIVED"
	StatusApproved Status = "APPROVED"
	StatusSeparated Status = "SEPARATED"
	StatusInRoute Status = "IN_ROUTE"
	StatusDelivered Status = "DELIVERED"
	StatusCancelled Status = "CANCELLED"
	StatusRescheduled Status = "RESCHEDULED"
)

var ErrInvalidTransition = errors.New("invalid status transition")

var transitions = map[Status][]Status{
	StatusReceived:  {StatusApproved, StatusCancelled},
	StatusApproved:  {StatusSeparated, StatusCancelled},
	StatusSeparated: {StatusInRoute, StatusCancelled},
	StatusInRoute:   {StatusDelivered, StatusCancelled},
	StatusDelivered: {},
	StatusCancelled: {},
	StatusRescheduled: {StatusApproved, StatusCancelled},
}

func CanTransitionTo(from, to Status) error {
  allowed, ok := transitions[from]
	if !ok {
		return ErrInvalidTransition
	}
	for _, s := range allowed {
		if s == to {
			return nil
		}
	}
	return ErrInvalidTransition
}

func isTerminal(s Status) bool {
	return len(transitions[s]) == 0
}

func AllStatuses() []Status {
	return []Status{
		StatusReceived,
		StatusApproved,
		StatusSeparated,
		StatusInRoute,
		StatusDelivered,
		StatusCancelled,
		StatusRescheduled,
	}
}

