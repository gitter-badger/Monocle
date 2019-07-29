package esi

import (
	"fmt"
	"net/url"
)

func (e *Client) GetCharactersCharacterID(id uint64, etag string) (Response, error) {

	version := 4
	path := fmt.Sprintf("/v%d/characters/%d/", version, id)

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

func (e *Client) HeadCharactersCharacterID(id uint64) (Response, error) {
	version := 4
	path := fmt.Sprintf("/v%d/characters/%d/", version, id)

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
