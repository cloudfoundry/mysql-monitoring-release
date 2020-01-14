package notificationemailer

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Debugger interface {
	Debug(string, map[string]interface{})
}

type Client struct {
	notificationsDomain string
	skipSSLCertVerify   bool
	logger              Debugger
}

func NewClient(
	notificationsDomain string,
	skipSSLCertVerify bool,
	logger Debugger,
) *Client {
	return &Client{
		notificationsDomain: notificationsDomain,
		skipSSLCertVerify:   skipSSLCertVerify,
		logger:              logger,
	}
}

type notificationsEmailRequestBody struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	HTML    string `json:"html"`
	KindID  string `json:"kind_id"`
}

func (c *Client) Email(clientToken, to, subject, html, kindID string) error {
	body, err := json.Marshal(notificationsEmailRequestBody{
		To:      to,
		Subject: subject,
		HTML:    html,
		KindID:  kindID,
	})
	if err != nil {
		c.logger.Debug("Error marshalling POST body to json", map[string]interface{}{
			"errMessage": err.Error(),
		})
		// Untested because we cannot get Marshal to fail - none of the fields
		// are un-marshallable
		return err
	}

	url := c.notificationsDomain + "/emails"
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		c.logger.Debug("Error creating new request to notifications", map[string]interface{}{
			"errorMessage": err.Error(),
		})
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("X-NOTIFICATIONS-VERSION", "1")
	req.Header.Add("Authorization", "Bearer "+clientToken)

	client := http.Client{
		Timeout: time.Second * 5,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: c.skipSSLCertVerify,
			},
		},
	}

	c.logger.Debug("Making request to notifications", map[string]interface{}{
		"method": req.Method,
		"url":    req.URL,
	})

	res, err := client.Do(req)
	if err != nil {
		c.logger.Debug("Error making request to notifications", map[string]interface{}{
			"errorMessage": err.Error(),
		})

		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusBadRequest {
		c.logger.Debug("received bad status code from notifications", map[string]interface{}{
			"statusCode": res.StatusCode,
		})

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.logger.Debug("Error reading response body", map[string]interface{}{
				"errorMessage": err.Error(),
			})
		}

		return fmt.Errorf("bad response %s (%d) - %s", "sending email", res.StatusCode, string(b))
	}

	return nil
}
