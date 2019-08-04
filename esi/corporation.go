package esi

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ddouglas/monocle"
	"github.com/pkg/errors"
)

func (e *Client) GetCorporationsCorporationID(corporation monocle.Corporation) (Response, error) {

	path := fmt.Sprintf("/v4/corporations/%d/", corporation.ID)

	url := url.URL{
		Scheme: "https",
		Host:   e.Host,
		Path:   path,
	}

	headers := make(map[string]string)

	if corporation.Etag != "" {
		headers["If-None-Match"] = corporation.Etag
	}

	request := Request{
		Method:  "GET",
		Path:    url,
		Headers: headers,
		Body:    []byte(""),
	}

	response, err := e.Request(request)
	if err != nil {
		return response, err
	}

	mx.Lock()
	e.Reset = RetrieveErrorResetFromResponse(response)
	e.Remain = RetrieveErrorCountFromResponse(response)
	mx.Unlock()

	switch response.Code {
	case 200:
		err := json.Unmarshal(response.Data.([]byte), &corporation)
		if err != nil {
			err = errors.Wrap(err, "unable to unmarshel response body")
			return response, err
		}
		expires, err := RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrap(err, "Error Encountered attempting to parse expires header")
			return response, err
		}
		corporation.Expires = expires

		etag, err := RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrap(err, "Error Encountered attempting to retrieve etag header")
			return response, err
		}
		corporation.Etag = etag

		break
	case 304:
		expires, err := RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrap(err, "Error Encountered attempting to parse expires header")
			return response, err
		}
		corporation.Expires = expires

		etag, err := RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrap(err, "Error Encountered attempting to retrieve etag header")
			return response, err
		}
		corporation.Etag = etag

		break
	case 500, 502, 503, 504:
		break
	default:
		err = errors.Wrapf(err, "Bad Response Code %d received from ESI API for url %s", response.Code, response.Path)
	}

	response.Data = corporation

	return response, err

}
