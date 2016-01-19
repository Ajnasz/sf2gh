package main

import (
	"net/url"
	"strconv"
)

type SFTicketResponse struct {
	SFTicket `json:"ticket"`
}

type SFMilestone struct {
	Closed   int  `json:"closed"`
	Complete bool `json:"complete"`
	// Default     bool   `json:"default"` // bool or string in sourceforge
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	Name        string `json:"name"`
	Total       int    `json:"total"`
}

type SFTicketListTicket struct {
	Summary   string `json:"summary"`
	TicketNum int    `json:"ticket_num"`
}

type SFTickets struct {
	Count      int                  `json:"count"`
	Limit      int                  `json:"limit"`
	Page       int                  `json:"page"`
	Milestones []SFMilestone        `json:"milestones"`
	Tickets    []SFTicketListTicket `json:"tickets"`
}

const ticketsPageLimit = 25

func GetSFTickets(project string, category string, page int) SFTickets {
	var tickets SFTickets

	values := url.Values{}

	values.Set("page", strconv.Itoa(page))
	values.Set("limit", strconv.Itoa(ticketsPageLimit))

	CallSFAPI(project, category, values, &tickets)

	return tickets
}
