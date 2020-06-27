package agent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type StatusResponse struct {
	Msg                      string `json:"status"`
	Time                     int64  `json:"time"`
	LastSuccessfulBackupTime int64  `json:"last_successful_backup_time"`
}

type Agent struct {
	client HTTPClient
}

func New(client HTTPClient) Agent {
	return Agent{client: client}
}

func (a Agent) Status(agentURL string) (string, int64, int64, error) {
	url := fmt.Sprintf("%s/status", agentURL)
	req, err := http.NewRequest("GET", url, ioutil.NopCloser(strings.NewReader("")))
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to make status http request: %v", err)
	}

	response, err := a.client.Do(req)
	if err != nil {
		return "", 0, 0, fmt.Errorf(`failed to get status at "%s": %v`, url, err)
	}

	if response.StatusCode != http.StatusOK {
		return "", 0, 0, fmt.Errorf("failed to get status, agent http response error: %d", response.StatusCode)
	}

	statusMsg := new(StatusResponse)
	if err = json.NewDecoder(response.Body).Decode(statusMsg); err != nil {
		return "", 0, 0, fmt.Errorf(`failed to unmarshal response from agent: %v`, err)
	}

	return statusMsg.Msg, statusMsg.Time, statusMsg.LastSuccessfulBackupTime, nil
}
