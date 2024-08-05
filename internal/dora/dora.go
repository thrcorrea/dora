package dora

import (
	"fmt"
	"sync"
	"time"

	"github.com/thrcorrea/dora/pkg/github"
	"github.com/thrcorrea/dora/pkg/shortcut"
)

var projects = []string{
	"sophia",
	"auth-people-front",
	"auth-people",
	"core",
	"envoy",
	"saati_api",
	"infra-prod",
	"infra-qa",
	"shortner",
	"kpis-automation",
}

type doraService struct {
	githubClient   github.GithubClient
	shortcutClient shortcut.ShortcutClient
}

type DoraService interface {
	GetDeploymentMetric(period []time.Time) (int, error)
	GetCFRMetric(period []time.Time, deploymentFrequency int) (int, error)
	GetCycleTime(period []time.Time) (string, error)
	GetMTTRMetric(period []time.Time) (string, error)
}

func NewDoraService(githubClient github.GithubClient, shortcutClient shortcut.ShortcutClient) DoraService {
	return &doraService{
		githubClient:   githubClient,
		shortcutClient: shortcutClient,
	}
}

func (d doraService) GetMTTRMetric(period []time.Time) (string, error) {
	stories, err := d.shortcutClient.ListStories(period, "bug")
	if err != nil {
		return "", err
	}

	cycleTime := 0
	for _, story := range stories {
		cycleTime += story.CycleTime
	}

	cycleTime = cycleTime / len(stories)
	days := cycleTime / 86400
	hours := (cycleTime % 86400) / 3600
	return fmt.Sprintf("%d Dias e %dh\n", days, hours), nil
}

func (d doraService) GetCycleTime(period []time.Time) (string, error) {
	stories, err := d.shortcutClient.ListStories(period, "")
	if err != nil {
		return "", err
	}

	cycleTime := 0
	for _, story := range stories {
		cycleTime += story.CycleTime
	}

	cycleTime = cycleTime / len(stories)
	days := cycleTime / 86400
	hours := (cycleTime % 86400) / 3600
	return fmt.Sprintf("%d Dias e %dh\n", days, hours), nil
}

func (d doraService) GetDeploymentMetric(period []time.Time) (int, error) {
	prsList := []github.PullRequest{}
	wg := &sync.WaitGroup{}
	for _, project := range projects {
		wg.Add(1)
		go func(prList []github.PullRequest, project string, wg *sync.WaitGroup) {
			prs, err := d.githubClient.ListPullrequests(project, period)
			if err != nil {
				panic(err)
			}
			prsList = append(prsList, prs...)
			wg.Done()
		}(prsList, project, wg)
	}
	wg.Wait()

	fmt.Println("PRs")
	for _, pr := range prsList {
		fmt.Printf("%s -  %v\n", pr.Title, pr.Url)
	}
	fmt.Println("")
	return len(prsList), nil
}

func (d doraService) GetCFRMetric(period []time.Time, deploymentFrequency int) (int, error) {
	stories, err := d.shortcutClient.ListStories(period, "bug")
	if err != nil {
		return 0, err
	}

	fmt.Println("Stories")
	for _, story := range stories {
		fmt.Printf("%s - %v\n", story.Name, story.Link)
	}
	fmt.Println("")

	return len(stories), nil
}
