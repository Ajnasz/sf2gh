package main

import (
	"bytes"
	"text/template"
	"time"

	"github.com/Ajnasz/sfapi"
)

// TicketFormatterData Stores information to create Github body
type TicketFormatterData struct {
	SFTicket *sfapi.Ticket
	Project  string
	Category string
	Imported time.Time
}

const formatTemplate = `Imported from SourceForge on {{.Imported | formatDate "2006-01-02 15:04"}}
Created by **{{.SFTicket.ReportedBy}}** on {{.SFTicket.CreatedDate}}
Original: https://sourceforge.net/p/{{.Project}}/{{.Category}}/{{.SFTicket.TicketNum}}

{{.SFTicket.Description}}

{{ if (gt (len .SFTicket.Attachments) 0) }}Attachments:
{{ range .SFTicket.Attachments}}- {{.URL}}
{{ end }}
{{ end }}
`

// FormatTicket Generates Github ticket body
func formatTicket(templateString string, ticket TicketFormatterData) (string, error) {
	funcMap := template.FuncMap{
		"formatDate": func(format string, t time.Time) string {
			return t.Format(format)
		},
	}
	tpl, err := template.New("ticket").Funcs(funcMap).Parse(templateString)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	tpl.Execute(&buf, ticket)

	return buf.String(), nil
}
