Imported from SourceForge on {{.Imported | formatDate "2006-01-02 15:04"}}
Created by **{{.SFTicket.ReportedBy}}** on {{.SFTicket.CreatedTime | formatDate "2006-01-02 15:04"}}
Original: https://sourceforge.net/p/{{.Project}}/{{.Category}}/{{.SFTicket.TicketNum}}

---

{{.SFTicket.Description}}

{{ if (gt (len .SFTicket.Attachments) 0) }}Attachments:
{{ range .SFTicket.Attachments}}- {{.URL}}
{{ end }}
{{ end }}
