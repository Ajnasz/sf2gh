package main

import (
	// "fmt"
	"log"
	"strings"
	"time"
)

func getPatchLabels(currentLabels []string, status string) []string {
	statusLabels := strings.Split(status, "-")

	return append(currentLabels, statusLabels[1:]...)
}

func getStatusText(ticket SFTicket) string {
	if strings.Split(ticket.Status, "-")[0] == "closed" {
		log.Println("Status closed")
		return "closed"
	} else {
		log.Println("Status open")
		return "open"
	}
}

func sfTicketToGhIssue(sfTicket SFTicket) {
	var ghTicket GHTicket

	log.Println("SF ticket data: ", sfTicket.Summary, sfTicket.TicketNum, sfTicket.Status)

	ghapi := CreateGHIssue("Ajnasz", "gh-api-test")

	ghapi.Create(GHIssueBody{
		Title: sfTicket.Summary,
		Body:  sfTicket.Description,
		// Assignee: sfTicket.AssignedTo,
		// Labels:   append(sfTicket.Labels, "bugs"),
	}, &ghTicket)

	ghpatcher := CreateGHExistingIssue("Ajnasz", "gh-api-test", ghTicket.Number)

	var editContainer GHTicket

	ghpatcher.Edit(GHIssueBody{
		Labels: getPatchLabels(append(sfTicket.Labels, "bugs"), sfTicket.Status),
		State:  getStatusText(sfTicket),
	}, &editContainer)

	log.Printf("ticket %d moved to %d\n", sfTicket.TicketNum, ghTicket.Number)
}

func main() {
	// ghapi := CreateGHIssue("Ajnasz", "gh-api-test")

	// var container interface{}
	// ghapi.Create(GHIssueBody{
	// 	Title:    "Test ticket title",
	// 	Body:     "Test ticket body",
	// 	Assignee: "Ajnasz",
	// 	Labels: []string{
	// 		"foo",
	// 		"bar",
	// 		"baz",
	// 	},
	// }, &container)
	// log.Println(container)

	page := 0

	for {
		log.Printf("Get page: %d", page)
		sfTickets := GetSFTickets("bugs", page)

		if len(sfTickets.Tickets) == 0 {
			return
		}

		for _, ticket := range sfTickets.Tickets {
			ticketVerb := GetSFTicket("bugs", ticket.TicketNum)

			sfTicketToGhIssue(ticketVerb)

			time.Sleep(time.Second * 2)
		}

		page += 1
	}
}
