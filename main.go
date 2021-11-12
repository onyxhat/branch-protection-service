package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/v40/github"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()

	// handle flags
	pat := flag.String("token", "", "personal access token used to oauth")
	org := flag.String("org", "", "organization to listen to")
	flag.Parse()
	if *pat == "" {
		panic("personal access token input not provided")
	}
	if *org == "" {
		panic("organization input not provided")
	}

	// auth & client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *pat},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	for {
		listenAndProtect(ctx, client, *org)
		// sleep 1 min
		time.Sleep(1 * time.Minute)
	}

}

// listenAndProtect will list org events and check for default branch for protection status.
// if a branch is not protected, it will update default branch to protected and mention user in issue.
func listenAndProtect(ctx context.Context, cli *github.Client, org string) {
	// get org events
	e, err := orgEvents(ctx, cli, org)
	if err != nil {
		log.Fatal(err)
	}

	for i := range e {
		repoName := strings.Split(e[i].Repo, "/")
		if len(repoName) != 2 {
			log.Fatal("unable to determine repository name")
		}
		log.Println("Repo:", repoName[1])
		log.Println("User:", e[i].User)
		log.Println("Default Branch:", e[i].MasterBranch)

		_, _, err := cli.Repositories.GetBranchProtection(ctx, org, repoName[1], e[i].MasterBranch)
		if err != nil {
			// GetBranchProtection returns an error when used on an unprotected branch. There is an open issue on this:
			// https://github.com/google/go-github/issues/625
			// If this behavior changes, this logic will need to change.
			//
			// attempt to protect branch
			if err := protectBranch(ctx, cli, e[i].MasterBranch, repoName[1], org); err != nil {
				log.Fatal("unable to protect branch: ", err)
			}
			log.Println("branch protected for ", e[i].MasterBranch)

			// perform @mention to user in repository issue
			if err := createIssue(ctx, cli, org, repoName[1], e[i].MasterBranch, e[i].User); err != nil {
				log.Fatal("unable to create issue: ", err)
			}
			log.Println("issue created")
		}
	}
}

// createIssue opens an issue to notify the user that the main/master branch is protected
func createIssue(ctx context.Context, cli *github.Client, org, repo, branch, user string) error {
	issueTitle := "Branch Protected"
	issueBody := fmt.Sprintf("@%s Default branch of %s has been protected from force pushing & deletion", user, branch)
	issue := github.IssueRequest{
		Title:    &issueTitle,
		Body:     &issueBody,
		Assignee: &user,
	}

	_, _, err := cli.Issues.Create(ctx, org, repo, &issue)
	if err != nil {
		return err
	}

	return nil
}

// EventRefInfo contains event information from listening
type EventRefInfo struct {
	Repo         string // name of repository
	User         string // user to mention
	Ref          string `json:"ref,omitempty"`           // used to check if event has branch
	MasterBranch string `json:"master_branch,omitempty"` // name of master branch
}

// orgEvents lists events for an organization and returns a slice of EventRefInfo
func orgEvents(ctx context.Context, cli *github.Client, org string) ([]EventRefInfo, error) {
	var events []EventRefInfo
	var repoMap = make(map[string]int)

	opt := github.ListOptions{PerPage: 30, Page: 1}
	orgEvents, _, err := cli.Activity.ListEventsForOrganization(ctx, org, &opt)
	if err != nil {
		return events, err
	}

	for i := range orgEvents {
		repoName := *orgEvents[i].Repo.Name

		// check RawPayload for ref
		// desired = events where branch exists
		var refValue EventRefInfo
		if err := json.Unmarshal(*orgEvents[i].RawPayload, &refValue); err != nil {
			return events, err
		}
		if len(refValue.Ref) == 0 {
			continue
		}

		// ignore duplicates if count is 1
		if repoMap[repoName] == 1 {
			continue
		}
		refValue.Repo = repoName
		refValue.User = *orgEvents[i].Actor.Login

		repoMap[repoName] = 1
		events = append(events, refValue)
	}
	return events, nil
}

// protectBranch will update a branch to disable force push and branch deletion
func protectBranch(ctx context.Context, cli *github.Client, branch, repo, owner string) error {
	branchSetting := false
	// protect branch from force pushes & force deletions
	protected := github.ProtectionRequest{
		AllowForcePushes: &branchSetting,
		AllowDeletions:   &branchSetting,
	}
	_, _, err := cli.Repositories.UpdateBranchProtection(ctx, owner, repo, branch, &protected)
	if err != nil {
		return err
	}

	return nil
}
