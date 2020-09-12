Created by **{{ .SFComment.Author }}** on {{ .SFComment.TimestampTime | formatDate "2006-01-02 15:04" }}

---
{{ if (ne (printf "#%d %s" .SFTicket.TicketNum .SFTicket.Summary) .SFComment.Subject)}}
*{{ .SFComment.Subject }}*

{{ .SFComment.Text }}
{{ else }}
{{ .SFComment.Text }}
{{ end }}


{{ if (gt (len .SFTicket.Attachments) 0) }}Attachments:
{{ range .SFTicket.Attachments}}- {{.URL}}
{{ end }}
{{ end }}
