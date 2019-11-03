package esi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/pkg/errors"
)

func (e *Client) HeadCorporationsCorporationID(id uint) (Response, error) {

	path := fmt.Sprintf("/v4/corporations/%d/", id)

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

	response, err := e.Request(request)
	if err != nil {
		return response, err
	}

	mx.Lock()
	e.Reset = RetrieveErrorResetFromResponse(response)
	e.Remain = RetrieveErrorCountFromResponse(response)
	mx.Unlock()

	switch response.Code {
	case 200, 500, 502, 503, 504:
		break
	default:
		err = fmt.Errorf("Code: %d Request: %s %s", response.Code, request.Method, url.Path)
	}
	return response, err
}

func (e *Client) GetCorporationsCorporationID(corporation *monocle.Corporation) (Response, error) {

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
		var newCorp monocle.Corporation

		err := json.Unmarshal(response.Data.([]byte), &newCorp)
		if err != nil {
			return response, errors.Wrap(err, "unable to unmarshel response body")
		}

		newCorp.ID = corporation.ID

		newCorp.Expires, err = RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			return response, errors.Wrap(err, "Error Encountered attempting to parse expires header")
		}

		newCorp.Etag, err = RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			return response, errors.Wrap(err, "Error Encountered attempting to retrieve etag header")
		}

		corporation = &newCorp

		break
	case 304:
		corporation.Expires, err = RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			return response, errors.Wrap(err, "Error Encountered attempting to parse expires header")
		}

		corporation.Etag, err = RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			return response, errors.Wrap(err, "Error Encountered attempting to retrieve etag header")
		}

		break
	case 503:
		time.Sleep(time.Minute * 5)
		break
	case 500, 502, 504:
		break
	default:
		err = fmt.Errorf("Code: %d Request: %s %s", response.Code, request.Method, url.Path)
	}

	response.Data = corporation

	return response, err

}

func (e *Client) GetCorporationsCorporationIDAllianceHistory(etagResource *monocle.EtagResource) (Response, error) {

	var history []*monocle.CorporationAllianceHistory

	path := fmt.Sprintf("/v2/corporations/%d/alliancehistory/", etagResource.ID)

	url := url.URL{
		Scheme: "https",
		Host:   e.Host,
		Path:   path,
	}

	headers := make(map[string]string)

	if etagResource.Etag != "" {
		headers["If-None-Match"] = etagResource.Etag
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

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &history)
		if err != nil {
			err = errors.Wrapf(err, "unable to unmarshel response body for %d corporation history: %s", etagResource.ID, err)
			return response, err
		}

		etagResource.Expires, err = RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			return response, errors.Wrapf(err, "Error Encountered attempting to parse expires header for url %s: %s", response.Path, err)
		}

		etagResource.Etag, err = RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			return response, errors.Wrapf(err, "Error Encountered attempting to retrieve etag header for url %s: %s", response.Path, err)
		}

		break
	case 304:
		etagResource.Expires, err = RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			return response, errors.Wrapf(err, "Error Encountered attempting to parse expires header for url %s: %s", response.Path, err)
		}

		etagResource.Etag, err = RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			return response, errors.Wrapf(err, "Error Encountered attempting to retrieve etag header for url %s: %s", response.Path, err)
		}

		break
	case 500, 502, 503, 504:
		break
	default:
		err = fmt.Errorf("Code: %d Request: %s %s", response.Code, request.Method, url.Path)
	}

	response.Data = map[string]interface{}{
		"history": history,
		"etag":    etagResource,
	}

	return response, err
}
