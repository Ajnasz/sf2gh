package main

import (
	// "fmt"
	"fmt"
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

func debug(args ...interface{}) {
	if false {
		log.Println(args)
	}
}

func printf(s string, args ...interface{}) {
	if false {
		fmt.Printf(s+"\n", args...)
	}
}

func getPatchLabels(currentLabels []string, status string) []string {
	statusLabels := strings.Split(status, "-")

	return append(currentLabels, statusLabels[1:]...)
}

func getStatusText(ticket *SFTicket) string {
	if strings.Split(ticket.Status, "-")[0] == "closed" {
		debug("Status closed")
		return "closed"
	} else {
		debug("Status open")
		return "open"
	}
}

func createSFBody(sfTicket *SFTicket, category string) *string {
	importText := fmt.Sprintf("Imported from SourceForge on %s", time.Now().Format(time.UnixDate))
	createdText := fmt.Sprintf("Created by **%s** on %s", sfTicket.ReportedBy, sfTicket.CreatedDate)
	link := fmt.Sprintf("Original: http://sourceforge.net/p/%s/%s/%d", project, category, sfTicket.TicketNum)
	body := fmt.Sprintf("%s\n%s\n%s\n\n%s", importText, createdText, link, sfTicket.Description)

	if len(sfTicket.Attachments) > 0 {
		attachments := []string{}

		for _, attachment := range sfTicket.Attachments {
			attachments = append(attachments, attachment.URL)
		}

		body += fmt.Sprintf("\n\nAttachments: %s", strings.Join(attachments, "\n"))
	}

	return &body
}

func createSFCommentBody(post *SFDiscussionPost, sfTicket *SFTicket) *string {
	createdText := fmt.Sprintf("Created by **%s** on %s", post.Author, post.Timestamp)
	var body string

	if post.Subject != fmt.Sprintf("#%d %s", sfTicket.TicketNum, sfTicket.Summary) {
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

func addCommentsToIssue(sfTicket *SFTicket, issue *github.Issue) {

	if len(sfTicket.DiscussionThread.Posts) > 0 {
		for _, post := range sfTicket.DiscussionThread.Posts {
			comment, response, err := githubClient.Issues.CreateComment(config.Github.UserName, ghRepo, *issue.Number, &github.IssueComment{
				Body: createSFCommentBody(&post, sfTicket),
			})

			if err != nil {
				log.Fatal(err)
			}

			debug("comment", comment)
			debug("response", response)
			time.Sleep(time.Millisecond * sleepTime)
		}

		printf("Comments added: %d", len(sfTicket.DiscussionThread.Posts))
	}
}

func findMatchingMilestone(sfTicket *SFTicket) int {
	ms := sfTicket.CustomFields.Milestone

	for _, milestone := range ghMilestones {
		if *milestone.Title == ms {
			return *milestone.Number
		}
	}

	return 0
}

func sfTicketToGhIssue(sfTicket *SFTicket, category string) {

	labels := getPatchLabels(append(sfTicket.Labels, category, "sourceforge"), sfTicket.Status)
	mileStone := findMatchingMilestone(sfTicket)

	issue, response, err := githubClient.Issues.Create(config.Github.UserName, ghRepo, &github.IssueRequest{
		Title:     &sfTicket.Summary,
		Body:      createSFBody(sfTicket, category),
		Labels:    &labels,
		Milestone: &mileStone,
		// Assignee: &sfTicket.AssignedTo,
		// State: &statusText,
	})

	if err != nil {
		log.Fatal(err)
	}

	statusText := getStatusText(sfTicket)

	if statusText != *issue.State {
		issue, response, err = githubClient.Issues.Edit(config.Github.UserName, ghRepo, *issue.Number, &github.IssueRequest{
			State: &statusText,
		})

		if err != nil {
			log.Fatal(err)
		}

	}

	addCommentsToIssue(sfTicket, issue)

	debug("response", response)
	debug("issue", issue)
	printf("ticket %d moved to %d\n", sfTicket.TicketNum, *issue.Number)
}

func getMilestonStatusText(milestone *SFMilestone) string {
	if milestone.Complete {
		return "closed"
	}

	return "open"
}

func createMilestones(sfTickets *SFTickets) {
	debug(sfTickets.Milestones)
	for _, milestone := range sfTickets.Milestones {
		status := getMilestonStatusText(&milestone)
		milestone, response, err := githubClient.Issues.CreateMilestone(config.Github.UserName, ghRepo, &github.Milestone{
			Title:       &milestone.Name,
			Description: &milestone.Description,
			State:       &status,
		})

		if err != nil {
			log.Fatal(err)
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
		time.Sleep(time.Millisecond * sleepTime)
	}
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
}

func main() {
	// ghapi := CreateGHIssue("Ajnasz", ghRepo)

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
	// debug(container)

	page := 0

	var progress *pb.ProgressBar

	category := "bugs"

	for {

		printf("Get page: %d", page)
		sfTickets := GetSFTickets(project, "bugs", page)

		createMilestones(&sfTickets)
		getMilestones()

		if progress == nil {
			progress = pb.StartNew(sfTickets.Count)
		}

		if len(sfTickets.Tickets) == 0 {
			return
		}

		for _, ticket := range sfTickets.Tickets {
			ticketVerb := GetSFTicket(project, category, ticket.TicketNum)

			sfTicketToGhIssue(&ticketVerb, category)
			progress.Increment()

			time.Sleep(time.Millisecond * sleepTime)
		}

		page += 1
	}

	progress.FinishPrint("All tickets imported")
}
