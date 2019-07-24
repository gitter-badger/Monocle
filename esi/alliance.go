package esi

import (
	"fmt"
	"net/url"
)

func (e *Client) GetAlliances(etag string) (Response, error) {
	version := 1
	path := fmt.Sprintf("/v%d/alliances/", version)

	url := url.URL{
		Scheme: "https",
		Host:   e.Host,
		Path:   path,
	}

	headers := make(map[string]string)

	if etag != "" {
		headers["If-None-Match"] = etag
	}

	request := Request{
		Method:  "GET",
		Path:    url,
		Headers: headers,
		Body:    []byte(""),
	}

	return e.Request(request)
}

func (e *Client) GetAlliancesAllianceID(id uint, etag string) (Response, error) {

	version := 3
	path := fmt.Sprintf("/v%d/alliances/%d/", version, id)

	url := url.URL{
		Scheme: "https",
		Host:   e.Host,
		Path:   path,
	}

	headers := make(map[string]string)

	if etag != "" {
		headers["If-None-Match"] = etag
	}

	request := Request{
		Method:  "GET",
		Path:    url,
		Headers: headers,
		Body:    []byte(""),
	}

	return e.Request(request)
}

func (e *Client) GetAllianceMembersByID(id uint, etag string) (Response, error) {
	version := 1
	path := fmt.Sprintf("/v%d/alliances/%d/corporations/", version, id)

	url := url.URL{
		Scheme: "https",
		Host:   e.Host,
		Path:   path,
	}

	headers := make(map[string]string)

	if etag != "" {
		headers["If-None-Match"] = etag
	}

	request := Request{
		Method:  "GET",
		Path:    url,
		Headers: headers,
		Body:    []byte(""),
	}

	return e.Request(request)
}
