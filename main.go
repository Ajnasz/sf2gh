package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Ajnasz/config-validator"
	"github.com/Ajnasz/sfapi"
	"github.com/cheggaaa/pb/v3"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
	"modernc.org/kv"
)

const dbFile = "progressDB"

// wait after each github api call
var _sleepTime int

var cliConfig CliConfig

var ghMilestones []*github.Milestone
var config Config
var githubClient *github.Client
var sfClient *sfapi.Client
var stopped bool

func sleepTillRateLimitReset(rate github.Rate) {
	if rate.Reset.After(time.Now()) {
		wait := rate.Reset.Sub(time.Now())
		log.Println("rate limit reached, waiting", wait)
		time.Sleep(wait)
	}
}

func getPatchLabels(currentLabels []string, status string) []string {
	statusLabels := strings.Split(status, "-")

	newLabel := strings.Join(statusLabels[1:], "-")

	if newLabel != "" {
		return append(currentLabels, newLabel)
	}
	return currentLabels
}

func getStatusText(ticket *sfapi.Ticket) string {
	if strings.Split(ticket.Status, "-")[0] == "closed" {
		return "closed"
	}
	return "open"
}

func createSFBody(ticket *sfapi.Ticket, category string) *string {
	importText := fmt.Sprintf("Imported from SourceForge on %s", time.Now().Format(time.UnixDate))
	createdText := fmt.Sprintf("Created by **%s** on %s", ticket.ReportedBy, ticket.CreatedDate)
	link := fmt.Sprintf("Original: http://sourceforge.net/p/%s/%s/%d", cliConfig.project, category, ticket.TicketNum)
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

func addCommentToIssue(ctx context.Context, post sfapi.DiscussionPost, ticket *sfapi.Ticket, issue *github.Issue) (*github.IssueComment, error) {
	issueComment, response, err := githubClient.Issues.CreateComment(ctx, config.Github.UserName, cliConfig.ghRepo, *issue.Number, &github.IssueComment{
		Body: createSFCommentBody(&post, ticket),
	})

	if err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			sleepTillRateLimitReset(response.Rate)
		} else {
			return nil, err
		}
	}
	return issueComment, nil
}

func addCommentsToIssue(ctx context.Context, progressDB ProgressState, ticket *sfapi.Ticket, issue *github.Issue) error {
	if len(ticket.DiscussionThread.Posts) > 0 {
		for _, post := range ticket.DiscussionThread.Posts {
			if stopped {
				return nil
			}
			if _, found, _ := progressDB.Get("comment", post.Slug); found {
				// fmt.Println("Skip creating comment")
				continue
			}

			issueComment, err := addCommentToIssue(ctx, post, ticket, issue)

			if err != nil {
				return err
			}
			progressDB.Set("comment", post.Slug, uint64(*issueComment.ID))
			time.Sleep(time.Millisecond * cliConfig.sleepTime)
		}
	}

	return nil
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

func sfTicketToGhIssue(ctx context.Context, progressDB ProgressState, ticket *sfapi.Ticket, category string) (*github.Issue, error) {

	issueNumber, found, err := progressDB.Get("issue", ticket.ID)

	if err != nil {
		return nil, err
	}
	if found {
		// fmt.Println("Issue exists, querying")
		issue, _, err := githubClient.Issues.Get(ctx, config.Github.UserName, cliConfig.ghRepo, int(issueNumber))

		if err != nil {
			return nil, err
		}

		return issue, nil
	}

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

	issue, response, err := githubClient.Issues.Create(ctx, config.Github.UserName, cliConfig.ghRepo, &issueRequest)

	if err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			sleepTillRateLimitReset(response.Rate)
		} else {
			return nil, err
		}
	}

	statusText := getStatusText(ticket)

	if statusText != *issue.State {
		issue, response, err = githubClient.Issues.Edit(ctx, config.Github.UserName, cliConfig.ghRepo, *issue.Number, &github.IssueRequest{
			State: &statusText,
		})

		if err != nil {
			if _, ok := err.(*github.RateLimitError); ok {
				sleepTillRateLimitReset(response.Rate)
			} else {
				return nil, err
			}
		}

	}

	progressDB.Set("issue", ticket.ID, uint64(*issue.Number))
	return issue, nil
}

func getMilestonStatusText(milestone *sfapi.Milestone) string {
	if milestone.Complete {
		return "closed"
	}

	return "open"
}

func isAlreadyExistError(respError *github.ErrorResponse) bool {
	return len(respError.Errors) == 1 && respError.Errors[0].Code == "already_exists"
}

func createMileStone(ctx context.Context, progressDB ProgressState, ms sfapi.Milestone) error {
	progressDB.Get("milestone", ms.Name)
	if _, found, _ := progressDB.Get("milestone", ms.Name); found {
		return nil
	}

	status := getMilestonStatusText(&ms)
	milestone, response, err := githubClient.Issues.CreateMilestone(ctx, config.Github.UserName, cliConfig.ghRepo, &github.Milestone{
		Title:       &ms.Name,
		Description: &ms.Description,
		State:       &status,
	})

	if err != nil {
		if errResp, ok := err.(*github.ErrorResponse); ok && isAlreadyExistError(errResp) {
			return nil
		}
		if _, ok := err.(*github.RateLimitError); ok {
			sleepTillRateLimitReset(response.Rate)
		} else {
			return err
		}
	}

	if *milestone.State != status {
		milestone, response, err = githubClient.Issues.EditMilestone(ctx, config.Github.UserName, cliConfig.ghRepo, *milestone.Number, &github.Milestone{
			State: &status,
		})

		if err != nil {
			if _, ok := err.(*github.RateLimitError); ok {
				sleepTillRateLimitReset(response.Rate)
			} else {
				return err
			}
		}
	}

	progressDB.Set("milestone", fmt.Sprint(*milestone.ID), uint64(*milestone.Number))
	return nil
}

func createMilestones(ctx context.Context, progressDB ProgressState, tickets *sfapi.TrackerInfo) error {
	log.Println("Creating milestones")

	progress := pb.StartNew(len(tickets.Milestones))
	for _, ms := range tickets.Milestones {
		if stopped {
			return nil
		}
		if err := createMileStone(ctx, progressDB, ms); err != nil {
			progress.Finish()
			return err
		}
		progress.Increment()

		time.Sleep(time.Millisecond * cliConfig.sleepTime)
	}

	progress.Finish()
	return nil
}

func getMilestones(ctx context.Context) error {
	milestones, response, err := githubClient.Issues.ListMilestones(ctx, config.Github.UserName, cliConfig.ghRepo, &github.MilestoneListOptions{
		State: "all",
	})

	if err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			sleepTillRateLimitReset(response.Rate)
		} else {
			return err
		}
	}

	ghMilestones = milestones
	return nil
}

func getFullSfTicket(category string, info sfapi.TrackerInfoTicket) (*sfapi.Ticket, error) {
	ticket, _, err := sfClient.Tracker.Get(category, info.TicketNum)

	return ticket, err
}

// ProgressItem is Struct to define progress
type ProgressItem struct{}

func init() {
	flag.IntVar(&_sleepTime, "sleepTime", 1550, "Sleep between api calls, github may stop you use the API if you call it too frequently")
	flag.StringVar(&cliConfig.ghRepo, "ghRepo", "", "Github repository name")
	flag.StringVar(&cliConfig.project, "project", "", "Sourceforge project")
	flag.Parse()

	cliConfig.sleepTime = time.Duration(_sleepTime)

	err := configValidator.Validate(cliConfig)

	if err != nil {
		log.Fatal(err)
	}
}

func getDB(dbFile string, opts *kv.Options) (*kv.DB, error) {
	createOpen := kv.Open
	status := "opening"

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		createOpen = kv.Create
		status = "creating"
	}

	if opts == nil {
		opts = &kv.Options{}
	}

	db, err := createOpen(dbFile, opts)

	if err != nil {
		return nil, fmt.Errorf("error %s %s: %v", status, dbFile, err)
	}

	return db, nil
}

func createTicket(ctx context.Context, progressDB ProgressState, category string, tk sfapi.TrackerInfoTicket) error {
	ticket, err := getFullSfTicket(category, tk)

	if err != nil {
		return err
	}

	issue, err := sfTicketToGhIssue(ctx, progressDB, ticket, category)
	if err != nil {
		return err
	}
	err = addCommentsToIssue(ctx, progressDB, ticket, issue)
	if err != nil {
		return err
	}

	return nil

}

func createTickets(ctx context.Context, progressDB ProgressState, tickets *sfapi.TrackerInfo, category string) (bool, error) {
	if len(tickets.Tickets) == 0 {
		return false, nil
	}
	progress := pb.StartNew(len(tickets.Tickets))
	log.Println(fmt.Sprintf("Creating tickets %d of %d", len(tickets.Tickets)+tickets.Page*tickets.Limit, tickets.Count))
	for _, ticket := range tickets.Tickets {
		if stopped {
			return false, nil
		}
		createTicket(ctx, progressDB, category, ticket)
		progress.Increment()
		time.Sleep(time.Millisecond * cliConfig.sleepTime)
	}

	progress.Finish()

	return true, nil
}

func doMigration(category string, progressDB ProgressState) {
	ctx := context.Background()
	var page int
	var limit int
	query := sfapi.NewRequestQuery()
	query.Limit = 100
	for {
		if stopped {
			return
		}
		query.Page = page
		tickets, _, err := sfClient.Tracker.Info(category, *query)
		page = tickets.Page
		limit = tickets.Limit

		if err != nil {
			log.Println(err)
			return
		}

		if ghMilestones == nil {
			err = createMilestones(ctx, progressDB, tickets)
			if err != nil {
				log.Println(err)
				return
			}
			err = getMilestones(ctx)
			if err != nil {
				log.Println(err)
				return
			}
		}

		if ok, err := createTickets(ctx, progressDB, tickets, category); !ok {
			if err != nil {
				log.Println(err)
			}
		}

		if page*limit+limit >= tickets.Count {
			break
		}

		page++
	}
}

func main() {
	config = GetConfig()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Github.AccessToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	githubClient = github.NewClient(tc)

	sfClient = sfapi.NewClient(nil, cliConfig.project)

	progressDB, err := CreateKVProgressState(dbFile)
	if err != nil {
		log.Fatal(err)
	}

	defer progressDB.Close()
	category := "bugs"

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT)

	go func() {
		<-signalChan
		stopped = true
		fmt.Println("Exiting")
	}()
	doMigration(category, progressDB)
}
