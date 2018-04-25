package mailinator

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httputil"
)

type InboxListResponse struct {
	Messages []InboxListMessage `json:"messages"`
}

type InboxListMessage struct {
	Subject string `json:"subject"`
	From    string `json:"from"`
	To      string `json:"to"`
}

type Client struct {
	APIToken string
}

func NewClient(token string) *Client {
	return &Client{
		APIToken: token,
	}
}

func (c *Client) GetMessageList(email string) (*InboxListResponse, error) {
	req, err := http.NewRequest("GET", "https://api.mailinator.com/api/inbox", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("token", c.APIToken)
	q.Add("to", email)
	req.URL.RawQuery = q.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode > 399 {
		b, _ := httputil.DumpResponse(res, true)

		return nil, errors.New("Unsuccessful Request: " + string(b))
	}

	var parsedResponse InboxListResponse

	json.NewDecoder(res.Body).Decode(&parsedResponse)

	return &parsedResponse, nil
}
