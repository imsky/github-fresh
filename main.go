package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// todo: docs
// todo: GitHub action
// todo: add minimum age of pull request
// todo: close old pull requests

// variables used to embed build information in the binary
var (
	BuildTime string
	BuildSHA  string
	Version   string
)

var crash = log.Fatalf

type pullRequest struct {
	Number    uint32    `json:"number"`
	UpdatedAt time.Time `json:"updated_at"`

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

// Executor provides runtime configuration and facilities
type Executor struct {
	client *http.Client
	token  string
	http   bool
	dry    bool
}

// NewExecutor returns a new executor of GitHub operations
func NewExecutor(token string, dry bool) *Executor {
	ex := Executor{
		client: &http.Client{},
		token:  token,
		dry:    dry,
	}

	return &ex
}

func (ex *Executor) makeRequest(method string, url string) (res *http.Response, err error) {
	protocol := "https"
	if ex.http {
		protocol = "http"
	}

	req, err := http.NewRequest(method, protocol+"://api.github.com/"+url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "token "+ex.token)
	res, err = ex.client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (ex *Executor) listClosedPullRequests(user string, repo string, days int) ([]pullRequest, error) {
	pullRequests := make([]pullRequest, 0, 1)
	now := time.Now()
	maxAgeHours := float64(days*24) + 0.01

	for page, keepGoing := 1, true; keepGoing; page++ {
		res, err := ex.makeRequest("GET", "repos/"+user+"/"+repo+"/pulls?state=closed&sort=updated&direction=desc&per_page=100&page="+strconv.Itoa(page))

		if err != nil {
			return pullRequests, errors.New("failed to get pull requests (" + err.Error() + ")")
		}

		d := json.NewDecoder(res.Body)
		var prs struct {
			PullRequests []pullRequest
		}
		err = d.Decode(&prs.PullRequests)

		if err != nil {
			return pullRequests, errors.New("failed to parse pull requests (" + err.Error() + ")")
		}

		for _, pr := range prs.PullRequests {
			prAge := now.Sub(pr.UpdatedAt).Hours()
			if prAge > maxAgeHours {
				keepGoing = false
				break
			}
			pullRequests = append(pullRequests, pr)
		}

		if len(prs.PullRequests) == 0 || len(prs.PullRequests) < 100 {
			break
		}
	}

	return pullRequests, nil
}

func (ex *Executor) listUnprotectedBranches(user string, repo string) ([]branch, error) {
	branches := make([]branch, 0, 1)

	for page := 1; ; page++ {
		res, err := ex.makeRequest("GET", "repos/"+user+"/"+repo+"/branches?protected=false&per_page=100&page="+strconv.Itoa(page))

		if err != nil {
			return branches, errors.New("failed to get branches (" + err.Error() + ")")
		}

		d := json.NewDecoder(res.Body)
		var bs struct {
			Branches []branch
		}
		err = d.Decode(&bs.Branches)

		if err != nil {
			return branches, errors.New("failed to parse branches (" + err.Error() + ")")
		}

		branches = append(branches, bs.Branches...)

		if len(bs.Branches) == 0 || len(bs.Branches) < 100 {
			break
		}
	}

	return branches, nil
}

func (ex *Executor) deleteBranches(user string, repo string, branches []string) (int, error) {
	deletedBranches := 0

	for _, branch := range branches {
		if ex.dry {
			log.Println("Will delete branch", branch)
			continue
		}

		_, err := ex.makeRequest("DELETE", "repos/"+user+"/"+repo+"/git/refs/heads/"+branch)

		if err != nil {
			return deletedBranches, errors.New("failed to delete branch " + branch + " (" + err.Error() + ")")
		}

		log.Println("Deleted branch", branch)
		deletedBranches++
	}

	return deletedBranches, nil
}

func getStaleBranches(branches []branch, pullRequests []pullRequest) []string {
	branchesByName := make(map[string]branch)
	staleBranches := make([]string, 0, 1)

	for _, b := range branches {
		branchesByName[b.Name] = b
	}

	for _, pr := range pullRequests {
		staleBranch, branchExists := branchesByName[pr.Head.Ref]
		if branchExists && staleBranch.Commit.SHA == pr.Head.SHA {
			staleBranches = append(staleBranches, pr.Head.Ref)
		}
	}

	return staleBranches
}

func getDays(days string) int {
	if days != "" {
		d, err := strconv.Atoi(days)
		if err == nil {
			return d
		}
	}
	return 1
}

func getDry(dry string) bool {
	retval, err := strconv.ParseBool(dry)
	if err == nil {
		return retval
	}
	return false
}

// Run finds branches of recently closed pull requests and deletes them
func Run(user string, repo string, days int, ex Executor) error {
	switch {
	case user == "":
		return errors.New("missing user")
	case repo == "":
		return errors.New("missing repo")
	case days < 1:
		return errors.New("invalid value for days (" + strconv.Itoa(days) + ")")
	}

	var err error

	closedPullRequests, err := ex.listClosedPullRequests(user, repo, days)
	if err != nil {
		return err
	}
	unprotectedBranches, err := ex.listUnprotectedBranches(user, repo)
	if err != nil {
		return err
	}
	staleBranches := getStaleBranches(unprotectedBranches, closedPullRequests)
	db, err := ex.deleteBranches(user, repo, staleBranches)
	if err != nil {
		return err
	}
	log.Println("Deleted " + strconv.Itoa(db) + " branches")
	return nil
}

func setupUsage() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "github-fresh v"+Version+" "+BuildTime+" "+BuildSHA+"\n\n")
		flag.PrintDefaults()
	}
}

func main() {
	var token = flag.String("token", os.Getenv("GITHUB_FRESH_TOKEN"), "GitHub API token (GITHUB_FRESH_TOKEN)")
	var user = flag.String("user", os.Getenv("GITHUB_FRESH_USER"), "GitHub user (GITHUB_FRESH_USER)")
	var repo = flag.String("repo", os.Getenv("GITHUB_FRESH_REPO"), "GitHub repo (GITHUB_FRESH_REPO)")
	var days = flag.Int("days", getDays(os.Getenv("GITHUB_FRESH_DAYS")), "Max age in days of checked pull requests (GITHUB_FRESH_DAYS)")
	var dry = flag.Bool("dry", getDry(os.Getenv("GITHUB_FRESH_DRY")), "Dry run (GITHUB_FRESH_DRY)")
	setupUsage()
	flag.Parse()
	if flag.NFlag() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	ex := NewExecutor(*token, *dry)
	err := Run(*user, *repo, *days, *ex)
	if err != nil {
		crash(err.Error())
	}
}
