package main

import (
	// "fmt"
	"fmt"
	"github.com/Ajnasz/sfapi"
	"github.com/cheggaaa/pb"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"log"
	"strings"
	"time"
)

// wait after each github api call
const sleepTime time.Duration = 1500

var ghRepo string
var ghMilestones []github.Milestone
var project string
var config Config
var githubClient *github.Client
var sfClient *sfapi.Client

func debug(args ...interface{}) {
	if false {
		log.Println(args...)
	}
}

func printf(s string, args ...interface{}) {
	if false {
		fmt.Printf(s+"\n", args...)
	}
}

func getPatchLabels(currentLabels []string, status string) []string {
	statusLabels := strings.Split(status, "-")

	newLabel := strings.Join(statusLabels[1:], "-")

	if newLabel != "" {
		return append(currentLabels, newLabel)
	} else {
		return currentLabels
	}
}

func getStatusText(ticket *sfapi.Ticket) string {
	if strings.Split(ticket.Status, "-")[0] == "closed" {
		debug("Status closed")
		return "closed"
	} else {
		debug("Status open")
		return "open"
	}
}

func createSFBody(ticket *sfapi.Ticket, category string) *string {
	importText := fmt.Sprintf("Imported from SourceForge on %s", time.Now().Format(time.UnixDate))
	createdText := fmt.Sprintf("Created by **%s** on %s", ticket.ReportedBy, ticket.CreatedDate)
	link := fmt.Sprintf("Original: http://sourceforge.net/p/%s/%s/%d", project, category, ticket.TicketNum)
	body := fmt.Sprintf("%s\n%s\n%s\n\n%s", importText, createdText, link, ticket.Description)

	if len(ticket.Attachments) > 0 {
		attachments := []string{}

		for _, attachment := range ticket.Attachments {
			attachments = append(attachments, attachment.URL)
		}

		body += fmt.Sprintf("\n\nAttachments: %s", strings.Join(attachments, "\n"))
	}

	return &body
}

func createSFCommentBody(post *sfapi.DiscussionPost, ticket *sfapi.Ticket) *string {
	createdText := fmt.Sprintf("Created by **%s** on %s", post.Author, post.Timestamp)
	var body string

	if post.Subject != fmt.Sprintf("#%d %s", ticket.TicketNum, ticket.Summary) {
		body = fmt.Sprintf("*%s*\n\n%s\n\n%s", post.Subject, createdText, post.Text)
	} else {
		body = fmt.Sprintf("%s\n\n%s", createdText, post.Text)
	}

	if len(post.Attachments) > 0 {
		attachments := []string{}

		for _, attachment := range post.Attachments {
			attachments = append(attachments, attachment.URL)
		}

		body += fmt.Sprintf("\n\nAttachments: %s", strings.Join(attachments, "\n"))
	}

	return &body
}

func addCommentsToIssue(ticket *sfapi.Ticket, issue *github.Issue) {

	if len(ticket.DiscussionThread.Posts) > 0 {
		for _, post := range ticket.DiscussionThread.Posts {
			comment, response, err := githubClient.Issues.CreateComment(config.Github.UserName, ghRepo, *issue.Number, &github.IssueComment{
				Body: createSFCommentBody(&post, ticket),
			})

			if err != nil {
				log.Fatal(err)
			}

			debug("comment", comment)
			debug("response", response)
			time.Sleep(time.Millisecond * sleepTime)
		}

		printf("Comments added: %d", len(ticket.DiscussionThread.Posts))
	}
}

func findMatchingMilestone(ticket *sfapi.Ticket) int {
	ms := ticket.CustomFields.Milestone

	for _, milestone := range ghMilestones {
		if *milestone.Title == ms {
			return *milestone.Number
		}
	}

	return 0
}

func sfTicketToGhIssue(ticket *sfapi.Ticket, category string) {

	labels := getPatchLabels(append(ticket.Labels, category, "sourceforge"), ticket.Status)
	mileStone := findMatchingMilestone(ticket)

	issueRequest := github.IssueRequest{
		Title:  &ticket.Summary,
		Body:   createSFBody(ticket, category),
		Labels: &labels,
		// Assignee: &ticket.AssignedTo,
		// State: &statusText,
	}

	if mileStone > 0 {
		issueRequest.Milestone = &mileStone
	}

	issue, response, err := githubClient.Issues.Create(config.Github.UserName, ghRepo, &issueRequest)

	if err != nil {
		log.Fatal(err)
	}

	statusText := getStatusText(ticket)

	if statusText != *issue.State {
		issue, response, err = githubClient.Issues.Edit(config.Github.UserName, ghRepo, *issue.Number, &github.IssueRequest{
			State: &statusText,
		})

		if err != nil {
			log.Fatal(err)
		}

	}

	addCommentsToIssue(ticket, issue)

	debug("response", response)
	debug("issue", issue)
	printf("ticket %d moved to %d\n", ticket.TicketNum, *issue.Number)
}

func getMilestonStatusText(milestone *sfapi.Milestone) string {
	if milestone.Complete {
		return "closed"
	}

	return "open"
}

func createMilestones(tickets *sfapi.TrackerInfo) {
	log.Println("Creating milestones")

	progress := pb.StartNew(len(tickets.Milestones))
	for _, milestone := range tickets.Milestones {
		status := getMilestonStatusText(&milestone)
		milestone, response, err := githubClient.Issues.CreateMilestone(config.Github.UserName, ghRepo, &github.Milestone{
			Title:       &milestone.Name,
			Description: &milestone.Description,
			State:       &status,
		})

		if err != nil {
			log.Println(err)
			continue
		}

		if *milestone.State != status {
			milestone, response, err = githubClient.Issues.EditMilestone(config.Github.UserName, ghRepo, *milestone.Number, &github.Milestone{
				State: &status,
			})

			if err != nil {
				log.Fatal(err)
			}
		}

		debug(milestone)
		debug(response)

		printf("Milestone %s created", *milestone.Title)

		progress.Increment()

		time.Sleep(time.Millisecond * sleepTime)
	}
	progress.FinishPrint("Milestones created")
}

func getMilestones() {
	milestones, response, err := githubClient.Issues.ListMilestones(config.Github.UserName, ghRepo, &github.MilestoneListOptions{
		State: "all",
	})

	if err != nil {
		log.Fatal(err)
	}

	debug(milestones)
	debug(response)
	ghMilestones = milestones
}

func init() {
	ghRepo = "gh-api-test"
	project = "fluxbox"
	config = GetConfig()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Github.AccessToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	githubClient = github.NewClient(tc)

	sfClient = sfapi.NewClient(nil, project)
}

func main() {

	var progress *pb.ProgressBar

	page := 0
	category := "bugs"

	for {

		printf("Get page: %d", page)
		tickets, _, err := sfClient.Tracker.Info(category)

		if err != nil {
			log.Fatal(err)
		}

		if ghMilestones == nil {
			createMilestones(tickets)
			getMilestones()
		}

		if progress == nil {
			log.Println("Creating tickets")
			progress = pb.StartNew(tickets.Count)
		}

		if len(tickets.Tickets) == 0 {
			break
		}

		for _, ticket := range tickets.Tickets {
			ticket, _, err := sfClient.Tracker.Get(category, ticket.TicketNum)

			if err != nil {
				log.Fatal(err)
			}

			sfTicketToGhIssue(ticket, category)

			progress.Increment()
			time.Sleep(time.Millisecond * sleepTime)
		}

		page += 1
	}

	progress.FinishPrint("All tickets imported")
}
