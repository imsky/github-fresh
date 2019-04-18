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

	PullRequestHead struct {
		Ref string `json:"ref"`
	} `json:"head"`
}

type PullRequestResult struct {
	PullRequests []PullRequest
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
	page := 1
	now := time.Now()
	maxAgeHours := float64(days) * 24

	for ; ; page++ {
		res, err := makeGithubAPIRequest("https://api.github.com/repos/"+user+"/"+repo+"/pulls?state=closed&sort=updated&direction=desc&per_page=100&page="+strconv.Itoa(page), "GET", token)

		if err != nil {
			panic(err)
		}

		decoder := json.NewDecoder(res.Body)
		var pullRequestResult PullRequestResult
		err = decoder.Decode(&pullRequestResult.PullRequests)

		if err != nil {
			panic(err)
		}

		if len(pullRequestResult.PullRequests) == 0 {
			break
		}

		lastPullRequest := pullRequestResult.PullRequests[len(pullRequestResult.PullRequests)-1]
		lastPullRequestAge := now.Sub(lastPullRequest.ClosedAt).Hours()

		pullRequests = append(pullRequests, pullRequestResult.PullRequests...)

		if lastPullRequestAge >= maxAgeHours {
			break
		}
	}

	return pullRequests
}

func main() {
	var token = flag.String("t", os.Getenv("GITHUB_TOKEN"), "GitHub API token")
	var user = flag.String("u", "", "GitHub user")
	var repo = flag.String("r", "", "GitHub repo")
	var days = flag.Int("d", 30, "Max age in days of checked pull requests")
	flag.Parse()
	fmt.Println(ListClosedPullRequests(*user, *repo, *days, *token))
	// u := url.URL{Host: "example.com", Path: "foo"}
}
