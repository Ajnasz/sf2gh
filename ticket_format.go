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

const ticketTemplate = `Imported from SourceForge on {{.Imported | formatDate "2006-01-02 15:04"}}
Created by **[{{.SFTicket.ReportedBy}}](https://sourceforge.net/u/{{.SFTicket.ReportedBy}}/)** on {{.SFTicket.CreatedTime | formatDate "2006-01-02 15:04"}}
Original: https://sourceforge.net/p/{{.Project}}/{{.Category}}/{{.SFTicket.TicketNum}}

---

{{.SFTicket.Description}}

{{ if (gt (len .SFTicket.Attachments) 0) }}Attachments:
{{ range .SFTicket.Attachments}}- {{.URL}}
{{ end }}
{{ end }}
`

type CommentFormatterData struct {
	Project   string
	Category  string
	Imported  time.Time
	SFComment *sfapi.DiscussionPost
	SFTicket  *sfapi.Ticket
}

const commentTemplate = `
Imported from SourceForge on {{.Imported | formatDate "2006-01-02 15:04"}}
Created by **[{{ .SFComment.Author }}](https://sourceforge.net/u/{{.SFComment.Author}}/)** on {{ .SFComment.TimestampTime | formatDate "2006-01-02 15:04" }}
Original: https://sourceforge.net/p/{{ .Project }}/{{ .Category }}/{{ .SFTicket.TicketNum }}/{{ .SFComment.Slug }}

---
{{ if (ne (printf "#%d %s" .SFTicket.TicketNum .SFTicket.Summary) .SFComment.Subject)}}
*{{ .SFComment.Subject }}*

{{ .SFComment.Text }}
{{ else }}
{{ .SFComment.Text }}
{{ end }}


{{ if (gt (len .SFComment.Attachments) 0) }}Attachments:
{{ range .SFComment.Attachments }}- {{ .URL }}
{{ end }}
{{ end }}
`

// FormatTicket Generates Github ticket body
func formatTpl(name string, templateString string, data interface{}) (string, error) {
	funcMap := template.FuncMap{
		"formatDate": func(format string, t time.Time) string {
			return t.Format(format)
		},
	}
	tpl, err := template.New(name).Funcs(funcMap).Parse(templateString)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	tpl.Execute(&buf, data)

	return buf.String(), nil
}

func FormatTicket(templateString string, data TicketFormatterData) (string, error) {
	return formatTpl("ticket", templateString, data)
}
func FormatComment(templateString string, data CommentFormatterData) (string, error) {
	return formatTpl("comment", templateString, data)
}
