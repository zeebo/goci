package heroku

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	api string
	app string
}

func New(app, api string) *Client {
	return &Client{
		api: api,
		app: app,
	}
}

type Process struct {
	Slug     string
	UPID     string
	Command  string
	Action   string
	Process  string
	Elapsed  int
	Attached bool
	State    string
}

func (c *Client) List() (p []*Process, err error) {
	lurl := fmt.Sprintf("https://api.heroku.com/apps/%s/ps", c.app)
	req, err := http.NewRequest("GET", lurl, nil)
	if err != nil {
		return
	}

	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth("", c.api)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&p)
	return
}

func (c *Client) Kill(ps string) (err error) {
	kurl := fmt.Sprintf("https://api.heroku.com/apps/%s/ps/stop", c.app)
	data := url.Values{
		"ps": {ps},
	}
	req, err := http.NewRequest("POST", kurl, strings.NewReader(data.Encode()))
	if err != nil {
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("", c.api)

	_, err = http.DefaultClient.Do(req)
	return
}

func (c *Client) Run(command string) (p *Process, err error) {
	rurl := fmt.Sprintf("https://api.heroku.com/apps/%s/ps", c.app)
	data := url.Values{
		"command": {command},
	}
	req, err := http.NewRequest("POST", rurl, strings.NewReader(data.Encode()))
	if err != nil {
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("", c.api)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	io.Copy(&buf, resp.Body)
	log.Println("run resp:", buf.String())

	dec := json.NewDecoder(&buf)
	err = dec.Decode(&p)
	return
}
