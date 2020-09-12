package main

import (
	"fmt"
	// "testing"
	"time"

	"github.com/Ajnasz/sfapi"
)

func ExampleFormatTicket_attachments() {
	t, _ := time.Parse(time.RFC3339Nano, "2020-09-11T18:38:20.999999999+02:00")

	vars := TicketFormatterData{
		SFTicket: &sfapi.Ticket{
			CreatedDate: "2001-08-14 16:50:45",
			TicketNum:   123,
			ReportedBy:  "reporter",
			Description: "Ticket Description",
			Attachments: []sfapi.TicketAttachment{
				{Bytes: 15, URL: "https://example.com/foo/bar/baz"},
				{Bytes: 15, URL: "https://example.com/qux/quux"},
			},
		},
		Project:  "project",
		Category: "cateory",
		Imported: t,
	}

	out, _ := formatTicket(formatTemplate, vars)

	fmt.Println(out)
	// Output:
	// Imported from SourceForge on 2020-09-11 18:38
	// Created by **reporter** on 2001-08-14 16:50:45
	// Original: https://sourceforge.net/p/project/cateory/123
	//
	// Ticket Description
	//
	//
	// Attachments:
	// - https://example.com/foo/bar/baz
	// - https://example.com/qux/quux
}
func ExampleFormatTicket_noattachments() {
	t, _ := time.Parse(time.RFC3339Nano, "2020-09-11T18:38:20.999999999+02:00")

	vars := TicketFormatterData{
		SFTicket: &sfapi.Ticket{
			CreatedDate: "2001-08-14 16:50:45",
			TicketNum:   123,
			ReportedBy:  "reporter",
			Description: "Ticket Description",
			Attachments: []sfapi.TicketAttachment{},
		},
		Project:  "project",
		Category: "cateory",
		Imported: t,
	}

	out, _ := formatTicket(formatTemplate, vars)

	fmt.Println(out)
	// Output:
	// Imported from SourceForge on 2020-09-11 18:38
	// Created by **reporter** on 2001-08-14 16:50:45
	// Original: https://sourceforge.net/p/project/cateory/123
	//
	// Ticket Description
}
