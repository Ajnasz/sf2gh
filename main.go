package main

import "log"

func main() {
	ghapi := CreateGHApi("/user/repos")

	var container interface{}
	ghapi.Get(&container)
	log.Println(container)
	// sfTickets := GetSFTickets("bugs")

	// for _, ticket := range sfTickets.Tickets {
	// 	ticketVerb := GetSFTicket("bugs", ticket.TicketNum)
	// 	log.Println(ticketVerb)
	// }
}
