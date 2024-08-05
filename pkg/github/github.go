package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type githubClient struct {
	token  string
	client *http.Client
}

type ListPullRequestsResponse struct {
	Url       string    `json:"html_url"`
	Title     string    `json:"title"`
	Merged_at time.Time `json:"merged_at"`
	Head      struct {
		Ref  string `json:"ref"`
		Repo struct {
			FullName string `json:"full_name"`
		} `json:"repo"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`
}

type PullRequest struct {
	Title      string
	RepoName   string
	Branch     string
	BaseBranch string
	Url        string
	MergedAt   time.Time
}

type GithubClient interface {
	ListPullrequests(repo string, period []time.Time) ([]PullRequest, error)
}

func NewGithubClient(token string) GithubClient {
	return &githubClient{
		token:  token,
		client: &http.Client{},
	}
}

func makeListPrRequest(repo string, page int, token string) ([]ListPullRequestsResponse, error) {
	requestUrl := fmt.Sprintf("https://api.github.com/repos/pin-people/%s/pulls?state=closed&base=main&page=%d&sort=updated&direction=desc", repo, page)
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	githubResponse := make([]ListPullRequestsResponse, 0)
	if res.StatusCode == http.StatusOK {
		err = json.Unmarshal(resBody, &githubResponse)
		if err != nil {
			return nil, err
		}
		return githubResponse, nil
	}

	return githubResponse, nil
}

func (gc githubClient) ListPullrequests(repo string, period []time.Time) ([]PullRequest, error) {
	periodStart := period[0]
	periodEnd := period[1]

	pullRequests := []PullRequest{}
	page := 0
	for outOfWindow := false; !outOfWindow; {
		page++
		prs, err := makeListPrRequest(repo, page, gc.token)
		if err != nil {
			return nil, err
		}
		if len(prs) == 0 {
			break
		}

		for _, pr := range prs {
			if !time.Time.IsZero(pr.Merged_at) {
				if pr.Merged_at.Before(periodStart) {
					outOfWindow = true
					break
				}

				if pr.Merged_at.After(periodStart) && pr.Merged_at.Before(periodEnd) {
					pullRequests = append(pullRequests, PullRequest{
						Title:      pr.Title,
						RepoName:   pr.Head.Repo.FullName,
						Branch:     pr.Head.Ref,
						BaseBranch: pr.Base.Ref,
						MergedAt:   pr.Merged_at,
						Url:        pr.Url,
					})
				}
			}
		}
	}

	return pullRequests, nil
}
