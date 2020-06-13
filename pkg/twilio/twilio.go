package twilio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/harrybrwn/errs"
)

var defaultClient *Client

const twilioHost = "api.twilio.com"

func init() {
	sid := os.Getenv("TWILIO_SID")
	token := os.Getenv("TWILIO_TOKEN")
	defaultClient = NewClient(sid, token)
}

// NewClient will create a new twilio client.
func NewClient(sid, token string) *Client {
	return ClientFromClient(sid, token, &http.Client{})
}

// ClientFromClient will create a twilio client that uses a user
// given http.Client.
func ClientFromClient(sid, token string, c *http.Client) *Client {
	return &Client{
		sid:    sid,
		token:  token,
		client: c,
	}
}

// Client is a twilio client
type Client struct {
	client *http.Client
	sid    string
	token  string

	SenderNumber string
}

// SetSender will set the default phone number used
// by the client's Send function.
func (c *Client) SetSender(phone string) {
	c.SenderNumber = phone
}

// SetSender will set the default phone number used
// by the client's Send function.
func SetSender(phone string) {
	defaultClient.SetSender(phone)
}

// Send will a message given the recipiant's phone number.
func (c *Client) Send(to, body string) (*MessageResponse, error) {
	if c.SenderNumber == "" {
		return nil, errors.New("could not find a SenderNumber")
	}
	return c.SendFrom(c.SenderNumber, to, body)
}

// Send will a message given the recipiant's phone number.
func Send(to, body string) (*MessageResponse, error) {
	return defaultClient.Send(to, body)
}

// SendFrom will send a message given the sender's number and the recipiant's number.
func (c *Client) SendFrom(from, to, body string) (*MessageResponse, error) {
	if c.sid == "" {
		return nil, errors.New("no twilio sid")
	}
	if c.token == "" {
		return nil, errors.New("no twilio token")
	}
	vals := url.Values{
		"To":   {to},
		"From": {from},
		"Body": {body},
	}
	resp, err := c.client.Do(c.newPostReq(
		path.Join("/2010-04-01/Accounts/", c.sid, "/Messages.json"),
		strings.NewReader(vals.Encode())),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		e := &Error{}
		return nil, errs.Pair(e, json.NewDecoder(resp.Body).Decode(e))
	}
	mr := &MessageResponse{}
	return mr, json.NewDecoder(resp.Body).Decode(mr)
}

// SendFrom will send a message given the sender's number and the recipiant's number.
func SendFrom(from, to, body string) (*MessageResponse, error) {
	return defaultClient.SendFrom(from, to, body)
}

// MessageResponse is the json response given from sending a message.
type MessageResponse struct {
	Sid                 string      `json:"sid"`
	DateCreated         string      `json:"date_created"`
	DateUpdated         string      `json:"date_updated"`
	DateSent            interface{} `json:"date_sent"`
	AccountSid          string      `json:"account_sid"`
	To                  string      `json:"to"`
	From                string      `json:"from"`
	MessagingServiceSid interface{} `json:"messaging_service_sid"`
	Body                string      `json:"body"`
	Status              string      `json:"status"`
	NumSegments         string      `json:"num_segments"`
	NumMedia            string      `json:"num_media"`
	Direction           string      `json:"direction"`
	APIVersion          string      `json:"api_version"`
	Price               interface{} `json:"price"`
	PriceUnit           string      `json:"price_unit"`
	ErrorCode           interface{} `json:"error_code"`
	ErrorMessage        interface{} `json:"error_message"`
	URI                 string      `json:"uri"`
	SubresourceUris     struct {
		Media string `json:"media"`
	} `json:"subresource_uris"`
}

func (c *Client) newPostReq(path string, body io.Reader) *http.Request {
	var (
		rc io.ReadCloser
		ok bool
	)
	if rc, ok = body.(io.ReadCloser); !ok {
		rc = ioutil.NopCloser(body)
	}
	req := &http.Request{
		Method: "POST",
		Body:   rc,
		URL: &url.URL{
			Scheme: "https",
			Host:   twilioHost,
			Path:   path,
		},
		Header: http.Header{
			"Accept":       {"application/json"},
			"Content-Type": {"application/x-www-form-urlencoded"},
		},
	}
	req.SetBasicAuth(c.sid, c.token)
	return req
}

// Error is a twilio response error
type Error struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	MoreInfo string `json:"more_info"`
	Status   int    `json:"status"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("twilio code %d: %s see %s", e.Code, e.Message, e.MoreInfo)
}
