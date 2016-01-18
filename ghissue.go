package main

import (
	"path"
	"strconv"
)

type GHTicket struct {
	Assignee struct {
		AvatarURL         string `json:"avatar_url"`
		EventsURL         string `json:"events_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		GravatarID        string `json:"gravatar_id"`
		HTMLURL           string `json:"html_url"`
		ID                int    `json:"id"`
		Login             string `json:"login"`
		OrganizationsURL  string `json:"organizations_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		ReposURL          string `json:"repos_url"`
		SiteAdmin         bool   `json:"site_admin"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		Type              string `json:"type"`
		URL               string `json:"url"`
	} `json:"assignee"`
	Body     string      `json:"body"`
	ClosedAt interface{} `json:"closed_at"`
	ClosedBy struct {
		AvatarURL         string `json:"avatar_url"`
		EventsURL         string `json:"events_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		GravatarID        string `json:"gravatar_id"`
		HTMLURL           string `json:"html_url"`
		ID                int    `json:"id"`
		Login             string `json:"login"`
		OrganizationsURL  string `json:"organizations_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		ReposURL          string `json:"repos_url"`
		SiteAdmin         bool   `json:"site_admin"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		Type              string `json:"type"`
		URL               string `json:"url"`
	} `json:"closed_by"`
	Comments    int    `json:"comments"`
	CommentsURL string `json:"comments_url"`
	CreatedAt   string `json:"created_at"`
	EventsURL   string `json:"events_url"`
	HTMLURL     string `json:"html_url"`
	ID          int    `json:"id"`
	Labels      []struct {
		Color string `json:"color"`
		Name  string `json:"name"`
		URL   string `json:"url"`
	} `json:"labels"`
	LabelsURL string `json:"labels_url"`
	Locked    bool   `json:"locked"`
	Milestone struct {
		ClosedAt     string `json:"closed_at"`
		ClosedIssues int    `json:"closed_issues"`
		CreatedAt    string `json:"created_at"`
		Creator      struct {
			AvatarURL         string `json:"avatar_url"`
			EventsURL         string `json:"events_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			GravatarID        string `json:"gravatar_id"`
			HTMLURL           string `json:"html_url"`
			ID                int    `json:"id"`
			Login             string `json:"login"`
			OrganizationsURL  string `json:"organizations_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			ReposURL          string `json:"repos_url"`
			SiteAdmin         bool   `json:"site_admin"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			Type              string `json:"type"`
			URL               string `json:"url"`
		} `json:"creator"`
		Description string `json:"description"`
		DueOn       string `json:"due_on"`
		HTMLURL     string `json:"html_url"`
		ID          int    `json:"id"`
		LabelsURL   string `json:"labels_url"`
		Number      int    `json:"number"`
		OpenIssues  int    `json:"open_issues"`
		State       string `json:"state"`
		Title       string `json:"title"`
		UpdatedAt   string `json:"updated_at"`
		URL         string `json:"url"`
	} `json:"milestone"`
	Number      int `json:"number"`
	PullRequest struct {
		DiffURL  string `json:"diff_url"`
		HTMLURL  string `json:"html_url"`
		PatchURL string `json:"patch_url"`
		URL      string `json:"url"`
	} `json:"pull_request"`
	State     string `json:"state"`
	Title     string `json:"title"`
	UpdatedAt string `json:"updated_at"`
	URL       string `json:"url"`
	User      struct {
		AvatarURL         string `json:"avatar_url"`
		EventsURL         string `json:"events_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		GravatarID        string `json:"gravatar_id"`
		HTMLURL           string `json:"html_url"`
		ID                int    `json:"id"`
		Login             string `json:"login"`
		OrganizationsURL  string `json:"organizations_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		ReposURL          string `json:"repos_url"`
		SiteAdmin         bool   `json:"site_admin"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		Type              string `json:"type"`
		URL               string `json:"url"`
	} `json:"user"`
}

type GHIssues struct {
	GHApi
}

type GHIssueBody struct {
	Title     string   `json:"title,omitempty"`
	Body      string   `json:"body,omitempty"`
	Assignee  string   `json:"assignee,omitempty"`
	Milestone int      `json:"milestone,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	State     string   `json:"state,omitempty"`
}

func CreateGHIssues() Api {
	return CreateGHApi("/issues")
}

func CreateGHUserIssues() Api {
	return CreateGHApi("/user/issues")
}

func CreateGHRepositoryIssues(user string, repo string) GHAPI {
	return CreateGHApi(path.Join("/repos", user, repo, "issues"))
}

func CreateGHIssue(user string, repo string) GHAPI {
	return CreateGHApi(path.Join("/repos", user, repo, "issues"))
}

func CreateGHExistingIssue(user string, repo string, id int) GHAPI {
	return CreateGHApi(path.Join("/repos", user, repo, "issues", strconv.Itoa(id)))
}
