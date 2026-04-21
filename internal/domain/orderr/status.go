package order

import "errors"

type Status string

const {
	StatusReceived Status = "RECEIVED"
	StatusApproved Status = "APPROVED"
	StatusSeparated Status = "SEPARATED"
	StatusInRoute Status = "IN_ROUTE"
	StatusDelivered Status = "DELIVERED"
	StatusCancelled Status = "CANCELLED"
	StatusRescheduled Status = "RESCHEDULED"
}

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
