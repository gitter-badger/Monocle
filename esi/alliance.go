package esi

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ddouglas/monocle"
	"github.com/pkg/errors"
)

func (e *Client) GetAlliances(etagResource monocle.EtagResource) (Response, monocle.EtagResource, error) {

	path := "/v1/alliances/"

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

	var ids = make([]int, 0)

	response, err := e.Request(request)
	if err != nil {
		return Response{}, etagResource, err
	}

	mx.Lock()
	e.Reset = RetrieveErrorResetFromResponse(response)
	e.Remain = RetrieveErrorCountFromResponse(response)
	mx.Unlock()

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &ids)
		if err != nil {
			err = errors.Wrap(err, "unable to unmarshal response body")
			return response, etagResource, err
		}

		expires, err := RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrapf(err, "Error Encounter with Request %s", path)

			return response, etagResource, err
		}

		etagResource.Expires = expires

		etag, err := RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrapf(err, "Error Encounter with Request %s", path)
			return response, etagResource, err
		}
		etagResource.Etag = etag
		break
	case 304:
		expires, err := RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrapf(err, "Error Encounter with Request %s", path)

			return response, etagResource, err
		}

		etagResource.Expires = expires

		etag, err := RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrapf(err, "Error Encounter with Request %s", path)
			return response, etagResource, err
		}
		etagResource.Etag = etag
		break

	default:
		err = errors.Wrapf(err, "Bad Response Code %d received from ESI API for url %s:", response.Code, response.Path)
		return response, etagResource, err
	}

	response.Data = ids

	return response, etagResource, err

}

func (e *Client) GetAlliancesAllianceID(alliance monocle.Alliance) (Response, error) {

	path := fmt.Sprintf("/v3/alliances/%d/", alliance.ID)

	url := url.URL{
		Scheme: "https",
		Host:   e.Host,
		Path:   path,
	}

	headers := make(map[string]string)

	if alliance.Etag != "" {
		headers["If-None-Match"] = alliance.Etag
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

	var updated monocle.Alliance

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &updated)
		if err != nil {
			err = errors.Wrapf(err, "unable to unmarshel response body on request %s", path)
			return response, err
		}

		expires, err := RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrapf(err, "Error Encounter with Request %s", path)

			return response, err
		}

		updated.Expires = expires

		etag, err := RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrapf(err, "Error Encounter with Request %s", path)
			return response, err
		}
		updated.Etag = etag
		break
	case 304:
		expires, err := RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrapf(err, "Error Encounter with Request %s", path)

			return response, err
		}

		updated.Expires = expires

		etag, err := RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrapf(err, "Error Encounter with Request %s", path)
			return response, err
		}
		updated.Etag = etag
		break
	default:
		err = fmt.Errorf("Bad Response Code %d received from ESI API for url %s:", response.Code, response.Path)
		return response, err
	}

	response.Data = updated

	return response, nil
}

func (e *Client) GetAllianceMembersByID(etagResource monocle.EtagResource) (Response, monocle.EtagResource, error) {

	path := fmt.Sprintf("/v1/alliances/%d/corporations/", etagResource.ID)

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

	var ids = make([]int, 0)

	response, err := e.Request(request)
	if err != nil {
		return response, etagResource, err
	}

	mx.Lock()
	e.Reset = RetrieveErrorResetFromResponse(response)
	e.Remain = RetrieveErrorCountFromResponse(response)
	mx.Unlock()

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &ids)
		if err != nil {
			err = errors.Wrapf(err, "unable to unmarshel response body on request %s", path)
			return response, etagResource, err
		}

		expires, err := RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrapf(err, "Error Encounter with Request %s", path)

			return response, etagResource, err
		}

		etagResource.Expires = expires

		etag, err := RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrapf(err, "Error Encounter with Request %s", path)
			return response, etagResource, err
		}
		etagResource.Etag = etag
		break
	case 304:
		expires, err := RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrapf(err, "Error Encounter with Request %s", path)

			return response, etagResource, err
		}

		etagResource.Expires = expires

		etag, err := RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrapf(err, "Error Encounter with Request %s", path)
			return response, etagResource, err
		}
		etagResource.Etag = etag
		break
	case 500, 502, 503, 504:
		break
	default:
		err = fmt.Errorf("Bad Response Code %d received from ESI API for url %s:", response.Code, response.Path)
	}

	response.Data = ids

	return response, etagResource, err
}
