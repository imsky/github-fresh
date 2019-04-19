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

// todo: get closed pull requests sorted by last updated time; stop processing after some cut-off (3h); get all unprotected branches - if any branches match those of the recently closed pull requests, delete them
// todo: long flags, flags and env vars for --token, --repo, optional --owner (implicitly check all repos)
// todo: Dockerfile, GitHub action

type PullRequest struct {
	Number   uint64    `json:"number"`
	ClosedAt time.Time `json:"closed_at"`

	Head struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"head"`
}

type PullRequestResult struct {
	PullRequests []PullRequest
}
type Branch struct {
	Name string `json:"name"`

	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

type BranchResult struct {
	Branches []Branch
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

func ListClosedPullRequests(user string, repo string, days int, token string) []PullRequest {
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
		var prr PullRequestResult
		err = d.Decode(&prr.PullRequests)

		if err != nil {
			panic(err)
		}

		if len(prr.PullRequests) == 0 {
			break
		}

		pullRequests = append(pullRequests, prr.PullRequests...)

		lastPullRequest := prr.PullRequests[len(prr.PullRequests)-1]
		lastPullRequestAge := now.Sub(lastPullRequest.ClosedAt).Hours()
		//todo: only add pull requests < maxAgeHours?
		if lastPullRequestAge >= maxAgeHours {
			break
		}
	}

	return pullRequests
}

func ListUnprotectedBranches(user string, repo string, token string) []Branch {
	branches := make([]Branch, 0, 1)

	for page := 1; ; page++ {
		res, err := makeGithubAPIRequest("https://api.github.com/repos/"+user+"/"+repo+"/branches?protected=false&per_page=100&page="+strconv.Itoa(page), "GET", token)

		if err != nil {
			panic(err)
		}

		d := json.NewDecoder(res.Body)
		var br BranchResult
		err = d.Decode(&br.Branches)

		if err != nil {
			panic(err)
		}

		if len(br.Branches) == 0 {
			break
		}

		branches = append(branches, br.Branches...)
	}

	return branches
}

func ListStaleBranches(closedPullRequests []PullRequest, branches []Branch) []string {
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

func DeleteBranches(user string, repo string, branches []string, token string) {
	for _, branch := range branches {
		_, err := makeGithubAPIRequest("https://api.github.com/repos/"+user+"/"+repo+"/git/refs/heads/"+branch, "DELETE", token)

		if err != nil {
			panic(err)
		}

		fmt.Println("Deleted ", branch)
	}
}

func run(user string, repo string, days int, token string) {
	closedPullRequests := ListClosedPullRequests(user, repo, days, token)
	unprotectedBranches := ListUnprotectedBranches(user, repo, token)
	staleBranches := ListStaleBranches(closedPullRequests, unprotectedBranches)
	fmt.Println(staleBranches)
	//DeleteBranches(user, repo, staleBranches, token)
	// u := url.URL{Host: "example.com", Path: "foo"}
}

func main() {
	var token = flag.String("t", os.Getenv("GITHUB_TOKEN"), "GitHub API token")
	var user = flag.String("u", "", "GitHub user")
	var repo = flag.String("r", "", "GitHub repo")
	var days = flag.Int("d", 30, "Max age in days of checked pull requests")
	flag.Parse()
	run(*user, *repo, *days, *token)
}
