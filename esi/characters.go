package esi

import (
	"fmt"
	"net/url"
)

func (e *Client) HeadCharactersCharacterID(id uint64) (Response, error) {

	path := fmt.Sprintf("/v4/characters/%d/", id)

	url := url.URL{
		Scheme: "https",
		Host:   e.Host,
		Path:   path,
	}

	headers := make(map[string]string)

	request := Request{
		Method:  "HEAD",
		Path:    url,
		Headers: headers,
		Body:    []byte(""),
	}

	return e.Request(request)
}

func (e *Client) GetCharactersCharacterID(id uint64, etag string) (Response, error) {

	path := fmt.Sprintf("/v4/characters/%d/", id)

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

func (e *Client) GetCharactersCharacterIDCorporationHistory(id uint64, etag string) (Response, error) {

	path := fmt.Sprintf("/v1/characters/%d/corporationhistory/", id)

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
