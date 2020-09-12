package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/signal"
	"sort"
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

var (
	_sleepTime int

	cliConfig CliConfig
	version   string
	build     string

	ghMilestones []*github.Milestone
	config       Config
	githubClient *github.Client
	sfClient     *sfapi.Client

	stopped               bool
	ticketTemplateString  string
	commentTemplateString string
)

var errExists = errors.New("Exists")
var categories = []string{"bugs", "patches", "feature-requests", "support-requests"}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

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

func createSFBody(ticket *sfapi.Ticket) (string, error) {
	return FormatTicket(ticketTemplateString, TicketFormatterData{
		SFTicket: ticket,
		Project:  cliConfig.project,
		Category: cliConfig.category,
		Imported: time.Now(),
	})
}

func createSFCommentBody(post *sfapi.DiscussionPost, ticket *sfapi.Ticket) (string, error) {
	return FormatComment(commentTemplateString, CommentFormatterData{
		Project:   cliConfig.project,
		Category:  cliConfig.category,
		Imported:  time.Now(),
		SFComment: post,
		SFTicket:  ticket,
	})
}

func addCommentToIssue(ctx context.Context, post sfapi.DiscussionPost, ticket *sfapi.Ticket, issue *github.Issue) (*github.IssueComment, error) {
	body, err := createSFCommentBody(&post, ticket)
	verbose("  Adding comment to issue")
	debug(fmt.Sprintf("Add comment to issue %+v", body))
	if err != nil {
		return nil, err
	}
	issueComment, response, err := githubClient.Issues.CreateComment(ctx, config.Github.UserName, cliConfig.ghRepo, *issue.Number, &github.IssueComment{
		Body: &body,
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
			if stopped && !cliConfig.skipComments {
				// do not exit in the middle of comments if skipComments is enabled
				return nil
			}

			if _, found, _ := progressDB.Get("comment", post.Slug); found {
				verbose("  Comment already exists, skipping")
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
		verbose("  Issue already exists, skipping")
		debug(fmt.Sprintf("Issue already created %+v", ticket))
		issue, _, err := githubClient.Issues.Get(ctx, config.Github.UserName, cliConfig.ghRepo, int(issueNumber))

		if err != nil {
			return nil, err
		}

		return issue, errExists
	}

	labels := getPatchLabels(append(ticket.Labels, category, "sourceforge"), ticket.Status)
	mileStone := findMatchingMilestone(ticket)

	body, err := createSFBody(ticket)

	if err != nil {
		return nil, err
	}

	issueRequest := github.IssueRequest{
		Title:  &ticket.Summary,
		Body:   &body,
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

	debug(fmt.Sprintf("Issue created: %+v", issue))

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
		debug("Skip creating milestone", ms.Name)
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

	debug("Milestone created", milestone)

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

	verbose(fmt.Sprintf("Creating Ticket: %s", ticket.Summary))
	debug(fmt.Sprintf("SF Ticket %+v", ticket))

	issue, err := sfTicketToGhIssue(ctx, progressDB, ticket, category)
	if err == errExists && !cliConfig.skipComments {
		// Ignore the error, go ahead with comment checks
	} else if err != nil {
		return err
	}

	debug(fmt.Sprintf("Ticket created %+v", issue))

	err = addCommentsToIssue(ctx, progressDB, ticket, issue)
	if err != nil {
		return err
	}

	return nil

}

func createTickets(ctx context.Context, progressDB ProgressState, tickets []sfapi.TrackerInfoTicket, category string) (bool, error) {
	progress := pb.StartNew(len(tickets))
	ts := &trackerInfoTicketSorter{tickets}
	sort.Sort(ts)
	for _, ticket := range tickets {
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

func getPagesCount(category string, expectedLimit int) (int, error) {
	debug("Get pages count")
	query := sfapi.NewRequestQuery()
	query.Limit = 1
	ticket, _, err := sfClient.Tracker.Info(category, *query)

	if err != nil {
		return 0, err
	}

	return int(math.Ceil(float64(ticket.Count / expectedLimit))), nil
}

func doMigration(category string, progressDB ProgressState) {
	debug("Start migration")
	ctx := context.Background()
	query := sfapi.NewRequestQuery()
	query.Limit = 10
	page, err := getPagesCount(category, query.Limit)
	debug("Pages", page)

	if err != nil {
		log.Println(err)
		return
	}

	for page >= 0 {
		if stopped {
			return
		}
		query.Page = page
		debug(fmt.Sprintf("Query tracker info. Category: %s, %+v", category, query))
		tickets, _, err := sfClient.Tracker.Info(category, *query)

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

		if len(tickets.Tickets) != 0 {
			log.Println(fmt.Sprintf("Creating tickets %d-%d of %d", tickets.Page*tickets.Limit, len(tickets.Tickets)+tickets.Page*tickets.Limit, tickets.Count))
			if ok, err := createTickets(ctx, progressDB, tickets.Tickets, category); !ok {
				if err != nil {
					log.Println(err)
				}
			}
		}

		page--
	}
	debug("Finish migration")
}

func init() {
	flag.IntVar(&_sleepTime, "sleepTime", 1550, "Sleep between api calls, github may stop you use the API if you call it too frequently")
	flag.StringVar(&cliConfig.ghRepo, "ghRepo", "", "Github repository name")
	flag.StringVar(&cliConfig.project, "project", "", "Sourceforge project")
	flag.StringVar(&cliConfig.ticketTemplate, "ticketTemplate", "", "Template file for a ticket")
	flag.StringVar(&cliConfig.commentTemplate, "commentTemplate", "", "Template file for a comments")
	flag.StringVar(&cliConfig.category, "category", "bugs", "Sourceforge category. One of: "+strings.Join(categories[:], ", "))
	flag.StringVar(&cliConfig.dbFile, "progressStorage", "", "File where the progress is stored, to ensure a ticket or comment is created only once")
	flag.BoolVar(&cliConfig.skipComments, "skipComments", false, "Do not check for new comments on already existing tickets")
	flag.BoolVar(&cliConfig.version, "version", false, "Show version and build information")
	flag.BoolVar(&cliConfig.verbose, "verbose", false, "Display more verbose progress")
	flag.BoolVar(&cliConfig.debug, "debug", false, "Display debug information")
}

func getTemplateString(defaultTemplate string, templateFileName string) (string, error) {
	if templateFileName != "" {
		data, err := ioutil.ReadFile(templateFileName)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return defaultTemplate, nil
}

func main() {
	flag.Parse()
	if cliConfig.dbFile == "" {
		cliConfig.dbFile = fmt.Sprintf("progress_%s.dat", cliConfig.ghRepo)
	}

	if cliConfig.version {
		fmt.Println(version, build)
		return
	}

	if !stringInSlice(cliConfig.category, categories) {
		fmt.Println("Category must be one of: " + strings.Join(categories[:], ", "))
		return
	}

	cliConfig.sleepTime = time.Duration(_sleepTime)

	err := configValidator.Validate(cliConfig)

	if err != nil {
		log.Fatal(err)
	}
	config = GetConfig()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Github.AccessToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	githubClient = github.NewClient(tc)

	sfClient = sfapi.NewClient(nil, cliConfig.project)

	progressDB, err := CreateKVProgressState(cliConfig.category, cliConfig.dbFile)
	if err != nil {
		log.Fatal(err)
	}

	defer progressDB.Close()
	ticketTemplateString, err = getTemplateString(ticketTemplate, cliConfig.ticketTemplate)
	if err != nil {
		log.Fatal(err)
	}
	commentTemplateString, err = getTemplateString(commentTemplate, cliConfig.commentTemplate)
	if err != nil {
		log.Fatal(err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT)

	go func() {
		<-signalChan
		stopped = true
		fmt.Println("\n*** Exiting due to user break")
		if cliConfig.skipComments {
			fmt.Println("*** Current ticket will be completed")
		}
	}()

	doMigration(cliConfig.category, progressDB)
}
