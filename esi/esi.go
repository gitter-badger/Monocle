package esi

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

var (
	LayoutESI = "Mon, 02 Jan 2006 15:04:05 MST"
)

type (
	// Client represents the application as a whole. Client has our HTTP Client, DB Client, and holds Secrets for Third Party API Communication

	Client struct {
		Host      string
		Http      *http.Client
		UserAgent string
	}
	Config struct {
		Host      string `envconfig:"ESI_HOST" required:"true"`
		UserAgent string `envconfig:"API_USER_AGENT" required:"true"`
	}

	Request struct {
		Method  string
		Path    url.URL
		Headers map[string]string
		Body    []byte
	}

	Response struct {
		Path    string
		Code    int
		Headers map[string]string
		Data    interface{}
	}
)

var err error

func New(prefix string) (*Client, error) {

	var config Config
	err = envconfig.Process(prefix, &config)
	if err != nil {
		return nil, err
	}

	http := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Client{
		Host:      config.Host,
		Http:      http,
		UserAgent: config.UserAgent,
	}, nil

}

func (e *Client) Request(request Request) (Response, error) {

	var rBody io.Reader

	if request.Body != nil {
		rBody = bytes.NewBuffer(request.Body)
	}

	req, err := http.NewRequest(request.Method, request.Path.String(), rBody)
	if err != nil {
		err = errors.Wrap(err, "Unable build request")
		return Response{}, err
	}
	for k, v := range request.Headers {
		req.Header.Add(k, v)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", e.UserAgent)

	resp, err := e.Http.Do(req)
	if err != nil {
		err = errors.Wrap(err, "Unable to make request")
		return Response{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.Wrap(err, "error reading body")
		return Response{}, err
	}

	var response Response
	response.Path = request.Path.String()
	response.Data = body
	response.Code = resp.StatusCode
	headers := make(map[string]string)
	for k, sv := range resp.Header {
		for _, v := range sv {
			headers[k] = v
		}
	}

	response.Headers = headers

	return response, nil
}

func RetreiveExpiresHeaderFromResponse(response Response) (time.Time, error) {
	if _, ok := response.Headers["Expires"]; !ok {
		err := fmt.Errorf("Expires Headers is missing for url %s", response.Path)
		return time.Time{}, err
	}
	expires, err := time.Parse(LayoutESI, response.Headers["Expires"])
	if err != nil {
		return expires, err
	}

	expires = expires.Add(time.Hour * 3)

	return expires, nil
}

func RetrieveEtagHeaderFromResponse(response Response) (string, error) {
	if _, ok := response.Headers["Etag"]; !ok {
		err = fmt.Errorf("Etag Header is missing from url %s", response.Path)
		return "", err
	}
	return response.Headers["Etag"], nil
}
