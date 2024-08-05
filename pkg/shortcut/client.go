package shortcut

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type shortcutClient struct {
	token  string
	client *http.Client
}

type ShortcutResponse struct {
	Next  string  `json:"next"`
	Data  []Story `json:"data"`
	Total int     `json:"total"`
}

type Story struct {
	Name        string    `json:"name"`
	Link        string    `json:"app_url"`
	Type        string    `json:"story_type"`
	CompletedAt time.Time `json:"completed_at"`
	CycleTime   int       `json:"cycle_time"`
}

type ShortcutClient interface {
	ListStories(completedPeriod []time.Time, storyType string) ([]Story, error)
}

func NewShortcutClient(token string) ShortcutClient {
	return &shortcutClient{
		token:  token,
		client: &http.Client{},
	}
}

func (s shortcutClient) makeListStoriesRequest(completedPeriod []time.Time, storyType string, next string, stories []Story) ([]Story, error) {
	requestUrl := ""
	baseUrl := "https://api.app.shortcut.com"
	if next != "" {
		requestUrl = baseUrl + next
	} else {

		query := fmt.Sprintf("completed:%s..%s", completedPeriod[0].Format("2006-01-02"), completedPeriod[1].Format("2006-01-02"))
		if storyType != "" {
			query = query + fmt.Sprintf("+type:%s", storyType)
		}
		requestUrl = fmt.Sprintf(`https://api.app.shortcut.com/api/v3/search/stories?query=%s&detail=slim`, query)
	}
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Shortcut-Token", s.token)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	shortcutResponse := ShortcutResponse{}
	if res.StatusCode == http.StatusOK {
		err = json.Unmarshal(resBody, &shortcutResponse)
		if err != nil {
			return nil, err
		}
		if shortcutResponse.Next != "" {
			return s.makeListStoriesRequest(completedPeriod, storyType, shortcutResponse.Next, append(stories, shortcutResponse.Data...))
		}
		return append(stories, shortcutResponse.Data...), nil
	} else {
		return nil, fmt.Errorf("Error fetching stories: %s", resBody)
	}
}

func (s shortcutClient) ListStories(completedPeriod []time.Time, storyType string) ([]Story, error) {
	return s.makeListStoriesRequest(completedPeriod, storyType, "", []Story{})
}
