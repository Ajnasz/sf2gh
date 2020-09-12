Imported from SourceForge on {{.Imported | formatDate "2006-01-02 15:04:05"}}
Created by **[{{ .SFComment.Author }}](https://sourceforge.net/u/{{.SFComment.Author}}/)** on {{ .SFComment.TimestampTime | formatDate "2006-01-02 15:04:05" }}
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
