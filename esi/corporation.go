package esi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/pkg/errors"
)

// HeadCorporationsCorporationID makes a HTTP GET Request to the /corporations/{corporation_id} endpoint
// Often used to see if a particular corporation exists or to check the remaining time until
// the cache expires
//
// Documentation: https://esi.evetech.net/ui/#/Corporation/get_corporations_corporation_id
// Version: v4
// Cache: 3600 sec (1 Hour)
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

	switch response.Code {
	case 200, 500, 502, 503, 504:
		break
	default:
		err = fmt.Errorf("Code: %d Request: %s %s", response.Code, request.Method, url.Path)
	}
	return response, err
}

// GetCorporationsCorporationID makes a HTTP GET Request to the /corporations/{corporation_id} endpoint
// for information about the provided corporation
//
// Documentation: https://esi.evetech.net/ui/#/Corporation/get_corporations_corporation_id
// Version: v4
// Cache: 3600 sec (1 Hour)
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

	switch response.Code {
	case 200:
		var newCorp monocle.Corporation

		err := json.Unmarshal(response.Data.([]byte), &newCorp)
		if err != nil {
			return response, errors.Wrap(err, "unable to unmarshel response body")
		}

		newCorp.ID = corporation.ID

		newCorp.Expires, err = RetrieveExpiresHeaderFromResponse(response, 0)
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
		corporation.Expires, err = RetrieveExpiresHeaderFromResponse(response, 0)
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

// GetCorporationsCorporationIDAllianceHistory makes a HTTP GET Request to the
// /v2/corporations/{corporation_id}/alliancehistory/ endpoint for a list of alliances that the
// provided corporation has previously been a member of.
//
// Documentation: https://esi.evetech.net/ui/#/Corporation/get_corporations_corporation_id_alliancehistory
// Version: v2
// Cache: 3600 sec (1 Hour)
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

		etagResource.Expires, err = RetrieveExpiresHeaderFromResponse(response, 0)
		if err != nil {
			return response, errors.Wrapf(err, "Error Encountered attempting to parse expires header for url %s: %s", response.Path, err)
		}

		etagResource.Etag, err = RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			return response, errors.Wrapf(err, "Error Encountered attempting to retrieve etag header for url %s: %s", response.Path, err)
		}

		break
	case 304:
		etagResource.Expires, err = RetrieveExpiresHeaderFromResponse(response, 0)
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
