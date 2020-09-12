package main

import (
	"github.com/Ajnasz/sfapi"
)

type trackerInfoTicketSorter struct {
	tickets []sfapi.TrackerInfoTicket
}

// Len is part of sort.Interface.
func (s *trackerInfoTicketSorter) Len() int {
	return len(s.tickets)
}

// Swap is part of sort.Interface.
func (s *trackerInfoTicketSorter) Swap(i, j int) {
	s.tickets[i], s.tickets[j] = s.tickets[j], s.tickets[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *trackerInfoTicketSorter) Less(i, j int) bool {
	return s.tickets[i].TicketNum < s.tickets[j].TicketNum
}
