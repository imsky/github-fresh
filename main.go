package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// todo: dry run
// todo: docs
// todo: comments
// todo: staticcheck
// todo: errcheck, structcheck, varcheck
// todo: test
// todo: GitHub action

var (
	BuildTime string
	BuildSHA  string
	Version   string
)

type pullRequest struct {
	Number   uint64    `json:"number"`
	ClosedAt time.Time `json:"closed_at"`

	Head struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"head"`
}

type branch struct {
	Name string `json:"name"`

	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

type githubAPI struct {
	client *http.Client
	token  string
}

func NewGitHubAPI(token string) *githubAPI {
	api := githubAPI{
		client: &http.Client{},
		token:  token,
	}

	return &api
}

func (api *githubAPI) makeRequest(method string, url string) (res *http.Response, err error) {
	req, err := http.NewRequest(method, "https://api.github.com/"+url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "token "+api.token)
	res, err = api.client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (api *githubAPI) listClosedPullRequests(user string, repo string, days int) []pullRequest {
	pullRequests := make([]pullRequest, 0, 1)
	now := time.Now()
	maxAgeHours := float64(days * 24)

	if maxAgeHours < 24.0 {
		maxAgeHours = 24.0
	}

	for page := 1; ; page++ {
		res, err := api.makeRequest("GET", "repos/"+user+"/"+repo+"/pulls?state=closed&sort=updated&direction=desc&per_page=100&page="+strconv.Itoa(page))

		if err != nil {
			log.Fatalln("Failed to get pull requests", err)
		}

		d := json.NewDecoder(res.Body)
		var prs struct {
			PullRequests []pullRequest
		}
		err = d.Decode(&prs.PullRequests)

		if err != nil {
			log.Fatalln("Failed to parse pull requests", err)
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

func (api *githubAPI) listUnprotectedBranches(user string, repo string) []branch {
	branches := make([]branch, 0, 1)

	for page := 1; ; page++ {
		res, err := api.makeRequest("GET", "repos/"+user+"/"+repo+"/branches?protected=false&per_page=100&page="+strconv.Itoa(page))

		if err != nil {
			log.Fatalln("Failed to get branches", err)
		}

		d := json.NewDecoder(res.Body)
		var bs struct {
			Branches []branch
		}
		err = d.Decode(&bs.Branches)

		if err != nil {
			log.Fatalln("Failed to parse branches", err)
		}

		if len(bs.Branches) == 0 {
			break
		}

		branches = append(branches, bs.Branches...)
	}

	return branches
}

func (api *githubAPI) deleteBranches(user string, repo string, branches []string) {
	for _, branch := range branches {
		_, err := api.makeRequest("DELETE", "repos/"+user+"/"+repo+"/git/refs/heads/"+branch)

		if err != nil {
			log.Fatalln("Failed to delete branch", branch, err)
		}

		log.Println("Deleted branch", branch)
	}
}

func listStaleBranches(closedPullRequests []pullRequest, branches []branch) []string {
	branchesByName := make(map[string]branch)
	staleBranches := make([]string, 0, 1)

	for _, b := range branches {
		branchesByName[b.Name] = b
	}

	for _, pr := range closedPullRequests {
		staleBranch, branchExists := branchesByName[pr.Head.Ref]
		if branchExists && staleBranch.Commit.SHA == pr.Head.SHA {
			staleBranches = append(staleBranches, pr.Head.Ref)
		}
	}

	return staleBranches
}

func getDays(envVar string) int {
	envDays := os.Getenv(envVar)
	if envDays != "" {
		d, err := strconv.Atoi(envDays)
		if err == nil {
			return d
		}
	}
	return 0
}

func Run(user string, repo string, days int, token string) ([]string, error) {
	var err error

	if user == "" {
		err = errors.New("Missing user")
	} else if repo == "" {
		err = errors.New("Missing repo")
	} else if days < 1 {
		err = errors.New("Invalid value for days:" + strconv.Itoa(days))
	} else if token == "" {
		err = errors.New("Missing token")
	}

	if err != nil {
		return []string{}, err
	}

	api := NewGitHubAPI(token)
	closedPullRequests := api.listClosedPullRequests(user, repo, days)
	unprotectedBranches := api.listUnprotectedBranches(user, repo)
	staleBranches := listStaleBranches(closedPullRequests, unprotectedBranches)
	api.deleteBranches(user, repo, staleBranches)
	return staleBranches, err
}

func main() {
	var token = flag.String("token", os.Getenv("GITHUB_FRESH_TOKEN"), "GitHub API token")
	var user = flag.String("user", os.Getenv("GITHUB_FRESH_USER"), "GitHub user")
	var repo = flag.String("repo", os.Getenv("GITHUB_FRESH_REPO"), "GitHub repo")
	var days = flag.Int("days", getDays("GITHUB_FRESH_DAYS"), "Max age in days of checked pull requests")
	flag.Parse()
	_, err := Run(*user, *repo, *days, *token)
	if err != nil {
		log.Fatalln(err)
	}
}
