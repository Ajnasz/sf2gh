package main

import "log"

func main() {
	sfTickets := GetSFTickets("bugs")

	for _, ticket := range sfTickets.Tickets {
		ticketVerb := GetSFTicket("bugs", ticket.TicketNum)
		log.Println(ticketVerb)
	}
}
