package esi

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	LayoutESI = "Mon, 02 Jan 2006 15:04:05 MST"
	err       error
	mx        sync.Mutex
)

type (
	// Client represents the application as a whole. Client has our HTTP Client, DB Client, and holds Secrets for Third Party API Communication

	Client struct {
		Host      string
		Http      *http.Client
		UserAgent string
		Remain    uint64 // Number of Error left until a 420 will be thrown
		Reset     uint64 // Number of Seconds remain until Remain is reset to 100
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

func New() (*Client, error) {

	http := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Client{
		Host:      viper.GetString("esi.host"),
		Http:      http,
		UserAgent: viper.GetString("api.user_agent"),
		Remain:    100,
		Reset:     60,
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

	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
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

func RetrieveExpiresHeaderFromResponse(response Response) (time.Time, error) {
	if _, ok := response.Headers["Expires"]; !ok {
		err := fmt.Errorf("Expires Headers is missing for url %s", response.Path)
		return time.Time{}, err
	}
	expires, err := time.Parse(LayoutESI, response.Headers["Expires"])
	if err != nil {
		return expires, err
	}

	expires = expires.Add(time.Hour * 12)

	return expires, nil
}

func RetrieveEtagHeaderFromResponse(response Response) (string, error) {
	if _, ok := response.Headers["Etag"]; !ok {
		err = fmt.Errorf("Etag Header is missing from url %s", response.Path)
		return "", err
	}
	return response.Headers["Etag"], nil
}

func RetrieveErrorCountFromResponse(response Response) uint64 {
	if _, ok := response.Headers["X-Esi-Error-Limit-Remain"]; !ok {
		err = fmt.Errorf("X-Esi-Error-Limit-Remain Header is missing from url %s", response.Path)
		return 100
	}

	count, err := strconv.ParseUint(response.Headers["X-Esi-Error-Limit-Remain"], 10, 32)
	if err != nil {
		return 100
	}

	return count
}

func RetrieveErrorResetFromResponse(response Response) uint64 {
	if _, ok := response.Headers["X-Esi-Error-Limit-Reset"]; !ok {
		err = fmt.Errorf("X-Esi-Error-Limit-Reset Header is missing from url %s", response.Path)
		return 100
	}
	seconds, err := strconv.ParseUint(response.Headers["X-Esi-Error-Limit-Reset"], 10, 32)
	if err != nil {
		return 100
	}

	return seconds
}
