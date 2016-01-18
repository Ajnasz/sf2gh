package main

import (
	"path"
	"strconv"
)

type SFTicket struct {
	_id          string        `json:"_id"`
	AssignedTo   string        `json:"assigned_to"`
	AssignedToID string        `json:"assigned_to_id"`
	Attachments  []interface{} `json:"attachments"`
	CreatedDate  string        `json:"created_date"`
	CustomFields struct {
		_milestone string `json:"_milestone"`
		_priority  string `json:"_priority"`
	} `json:"custom_fields"`
	Description        string `json:"description"`
	DiscussionDisabled bool   `json:"discussion_disabled"`
	DiscussionThread   struct {
		_id          string      `json:"_id"`
		DiscussionID string      `json:"discussion_id"`
		Limit        int         `json:"limit"`
		Page         interface{} `json:"page"`
		Posts        []struct {
			Attachments []struct {
				Bytes int    `json:"bytes"`
				URL   string `json:"url"`
			} `json:"attachments"`
			Author     string      `json:"author"`
			LastEdited interface{} `json:"last_edited"`
			Slug       string      `json:"slug"`
			Subject    string      `json:"subject"`
			Text       string      `json:"text"`
			Timestamp  string      `json:"timestamp"`
		} `json:"posts"`
		Subject string `json:"subject"`
	} `json:"discussion_thread"`
	DiscussionThreadURL string        `json:"discussion_thread_url"`
	Labels              []string      `json:"labels"`
	ModDate             string        `json:"mod_date"`
	Private             bool          `json:"private"`
	RelatedArtifacts    []interface{} `json:"related_artifacts"`
	ReportedBy          string        `json:"reported_by"`
	ReportedByID        string        `json:"reported_by_id"`
	Status              string        `json:"status"`
	Summary             string        `json:"summary"`
	TicketNum           int           `json:"ticket_num"`
	VotesDown           int           `json:"votes_down"`
	VotesUp             int           `json:"votes_up"`
}

func GetSFTicket(category string, id int) (ticket SFTicket) {
	var ticketResponse SFTicketResponse
	CallSFAPI(path.Join(category, strconv.Itoa(id)), nil, &ticketResponse)

	return ticketResponse.SFTicket
}
