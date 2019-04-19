package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

// todo: dry run
// todo: logging
// todo: error handling
// todo: docs
// todo: comments
// todo: cross-compile
// todo: lint and format
// todo: staticcheck
// todo: errcheck, structcheck, varcheck, go vet
// todo: test
// todo: Dockerfile, GitHub action

type PullRequest struct {
	Number   uint64    `json:"number"`
	ClosedAt time.Time `json:"closed_at"`

	Head struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"head"`
}

type Branch struct {
	Name string `json:"name"`

	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

func makeGithubAPIRequest(url string, method string, token string) (res *http.Response, err error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "token "+token)
	res, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func listClosedPullRequests(user string, repo string, days int, token string) []PullRequest {
	pullRequests := make([]PullRequest, 0, 1)
	now := time.Now()
	maxAgeHours := float64(days) * 24
	state := "closed"

	for page := 1; ; page++ {
		res, err := makeGithubAPIRequest("https://api.github.com/repos/"+user+"/"+repo+"/pulls?state="+state+"&sort=updated&direction=desc&per_page=100&page="+strconv.Itoa(page), "GET", token)

		if err != nil {
			panic(err)
		}

		d := json.NewDecoder(res.Body)
		var prs struct {
			PullRequests []PullRequest
		}
		err = d.Decode(&prs.PullRequests)

		if err != nil {
			panic(err)
		}

		if len(prs.PullRequests) == 0 {
			break
		}

		pullRequests = append(pullRequests, prs.PullRequests...)

		lastPullRequest := prs.PullRequests[len(prs.PullRequests)-1]
		lastPullRequestAge := now.Sub(lastPullRequest.ClosedAt).Hours()
		//todo: only add pull requests < maxAgeHours?
		if lastPullRequestAge >= maxAgeHours {
			break
		}
	}

	return pullRequests
}

func listUnprotectedBranches(user string, repo string, token string) []Branch {
	branches := make([]Branch, 0, 1)

	for page := 1; ; page++ {
		res, err := makeGithubAPIRequest("https://api.github.com/repos/"+user+"/"+repo+"/branches?protected=false&per_page=100&page="+strconv.Itoa(page), "GET", token)

		if err != nil {
			panic(err)
		}

		d := json.NewDecoder(res.Body)
		var bs struct {
			Branches []Branch
		}
		err = d.Decode(&bs.Branches)

		if err != nil {
			panic(err)
		}

		if len(bs.Branches) == 0 {
			break
		}

		branches = append(branches, bs.Branches...)
	}

	return branches
}

func listStaleBranches(closedPullRequests []PullRequest, branches []Branch) []string {
	branchMap := make(map[string]Branch)
	staleBranches := make([]string, 0, 1)

	for _, branch := range branches {
		branchMap[branch.Name] = branch
	}

	for _, pr := range closedPullRequests {
		staleBranch, branchExists := branchMap[pr.Head.Ref]
		if branchExists && staleBranch.Commit.SHA == pr.Head.SHA {
			staleBranches = append(staleBranches, pr.Head.Ref)
		}
	}

	return staleBranches
}

func deleteBranches(user string, repo string, branches []string, token string) {
	for _, branch := range branches {
		_, err := makeGithubAPIRequest("https://api.github.com/repos/"+user+"/"+repo+"/git/refs/heads/"+branch, "DELETE", token)

		if err != nil {
			panic(err)
		}

		fmt.Println("Deleted ", branch)
	}
}

func run(user string, repo string, days int, token string) {
	//todo: validate input
	closedPullRequests := listClosedPullRequests(user, repo, days, token)
	unprotectedBranches := listUnprotectedBranches(user, repo, token)
	staleBranches := listStaleBranches(closedPullRequests, unprotectedBranches)
	deleteBranches(user, repo, staleBranches, token)
	// u := url.URL{Host: "example.com", Path: "foo"}
}

func getDays() int {
	envDays := os.Getenv("GITHUB_FRESH_DAYS")
	if envDays != "" {
		d, err := strconv.Atoi(envDays)
		if err == nil {
			if d > 0 {
				return d
			}
		}
	}
	return 30
}

func main() {
	var token = flag.String("token", os.Getenv("GITHUB_FRESH_TOKEN"), "GitHub API token")
	var user = flag.String("user", os.Getenv("GITHUB_FRESH_USER"), "GitHub user")
	var repo = flag.String("repo", os.Getenv("GITHUB_FRESH_REPO"), "GitHub repo")
	var days = flag.Int("days", getDays(), "Max age in days of checked pull requests")
	flag.Parse()
	run(*user, *repo, *days, *token)
}
