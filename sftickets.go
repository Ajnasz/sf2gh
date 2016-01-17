package main

type SFTicketResponse struct {
	SFTicket `json:"ticket"`
}

type SFTicketListTicket struct {
	Summary   string `json:"summary"`
	TicketNum int    `json:"ticket_num"`
}

type SFTickets struct {
	Count   int                  `json:"count"`
	Limit   int                  `json:"limit"`
	Tickets []SFTicketListTicket `json:"tickets"`
}

func GetSFTickets(category string) SFTickets {
	var tickets SFTickets

	CallSFAPI(category, &tickets)

	return tickets
}
