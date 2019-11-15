package esi

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ddouglas/monocle"
	"github.com/pkg/errors"
)

// GetAlliancesAllianceID makes a HTTP HEAD Request to the /alliances/{alliance_id} endpoint
// Often used to see if a particular alliance exists or to check the remaining time until
// the cache expires
//
// Documentation: https://esi.evetech.net/ui/#/Alliance/get_alliances_alliance_id
// Version: v3
// Cache: 3600 sec (1 Hour)
func (e *Client) HeadAlliancesAllianceID(id uint) (Response, error) {

	path := fmt.Sprintf("/v3/alliances/%d/", id)

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

	switch response.Code {
	case 200, 500, 502, 503, 504:

		break

	default:
		err = fmt.Errorf("Code: %d Request: %s %s", response.Code, request.Method, url.Path)
	}

	return response, err
}

// GetAlliancesAllianceID makes a HTTP GET Request to the /alliances/{alliance_id} endpoint
// for information about the provided alliance
//
// Documentation: https://esi.evetech.net/ui/#/Alliance/get_alliances_alliance_id
// Version: v3
// Cache: 3600 sec (1 Hour)
func (e *Client) GetAlliancesAllianceID(alliance *monocle.Alliance) (Response, error) {

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
	switch response.Code {
	case 200:
		var newAlliance monocle.Alliance
		err = json.Unmarshal(response.Data.([]byte), &newAlliance)
		if err != nil {
			return response, errors.Wrapf(err, "unable to unmarshel response body on request %s", path)
		}

		newAlliance.ID = alliance.ID

		newAlliance.Expires, err = RetrieveExpiresHeaderFromResponse(response, 0)
		if err != nil {
			return response, errors.Wrapf(err, "Error Encounter with Request %s", path)
		}

		newAlliance.Etag, err = RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			return response, errors.Wrapf(err, "Error Encounter with Request %s", path)
		}

		alliance = &newAlliance

		break

	case 304:
		expires, err := RetrieveExpiresHeaderFromResponse(response, 0)
		if err != nil {
			err = errors.Wrapf(err, "Error Encounter with Request %s", path)

			return response, err
		}
		alliance.Expires = expires

		etag, err := RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrapf(err, "Error Encounter with Request %s", path)
			return response, err
		}
		alliance.Etag = etag

		break

	default:
		err = fmt.Errorf("Code: %d Request: %s %s", response.Code, request.Method, url.Path)
	}

	response.Data = alliance

	return response, err
}

// GetAlliancesAllianceIDCorporations makes a HTTP GET Request to the /alliances/{alliance_id}/corporations endpoint
// and receives a list of ids corresponding to the corporations that are currently a member of the provided
// alliance
//
// Documentation: https://esi.evetech.net/ui/#/Alliance/get_alliances_alliance_id_corporations
// Version: v1
// Cache: 3600 sec (1 Hour)
func (e *Client) GetAlliancesAllianceIDCorporations(etagResource *monocle.EtagResource) (Response, error) {

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

	var ids = make([]uint, 0)

	response, err := e.Request(request)
	if err != nil {
		return response, err
	}

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &ids)
		if err != nil {
			return response, errors.Wrapf(err, "unable to unmarshel response body on request %s", path)
		}

		etagResource.Expires, err = RetrieveExpiresHeaderFromResponse(response, 0)
		if err != nil {
			return response, errors.Wrapf(err, "Error Encounter with Request %s", path)
		}

		etagResource.Etag, err = RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			return response, errors.Wrapf(err, "Error Encounter with Request %s", path)
		}

		break

	case 304:
		etagResource.Expires, err = RetrieveExpiresHeaderFromResponse(response, 0)
		if err != nil {
			return response, errors.Wrapf(err, "Error Encounter with Request %s", path)
		}

		etagResource.Etag, err = RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			return response, errors.Wrapf(err, "Error Encounter with Request %s", path)
		}

		break

	case 500, 502, 503, 504:
		break
	default:
		err = fmt.Errorf("Code: %d Request: %s %s", response.Code, request.Method, url.Path)
	}

	response.Data = map[string]interface{}{
		"ids":  ids,
		"etag": etagResource,
	}

	return response, err
}
