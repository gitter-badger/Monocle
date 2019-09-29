package esi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/pkg/errors"
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

func (e *Client) GetCharactersCharacterID(character monocle.Character) (Response, error) {

	path := fmt.Sprintf("/v4/characters/%d/", character.ID)

	url := url.URL{
		Scheme: "https",
		Host:   e.Host,
		Path:   path,
	}

	headers := make(map[string]string)

	if character.Etag != "" {
		headers["If-None-Match"] = character.Etag
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
		var newChar monocle.Character
		err := json.Unmarshal(response.Data.([]byte), &newChar)
		if err != nil {
			err = errors.Wrap(err, "unable to unmarshel response body")
			return response, err
		}
		newChar.ID = character.ID

		if character.CorporationID == 1000001 {
			newChar.Ignored = true
		}

		expires, err := RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrap(err, "Error Encountered attempting to parse expires header")
			return response, err
		}
		newChar.Expires = expires

		etag, err := RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrap(err, "Error Encountered attempting to retrieve etag header")
			return response, err
		}
		newChar.Etag = etag

		character = newChar

		break
	case 304:
		expires, err := RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrap(err, "Error Encountered attempting to parse expires header")
			return response, err
		}
		character.Expires = expires

		etag, err := RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrap(err, "Error Encountered attempting to retrieve etag header")
			return response, err
		}
		character.Etag = etag

		break
	case 503:
		time.Sleep(time.Minute * 5)
		break
	case 500, 502, 504:
		break
	default:
		err = fmt.Errorf("Code: %d Request: %s %s", response.Code, request.Method, url.Path)
	}

	response.Data = character

	return response, err
}

func (e *Client) GetCharactersCharacterIDCorporationHistory(etagResource monocle.EtagResource) (Response, monocle.EtagResource, error) {

	var history []monocle.CharacterCorporationHistory

	path := fmt.Sprintf("/v1/characters/%d/corporationhistory/", etagResource.ID)

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
		return response, etagResource, err
	}

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &history)
		if err != nil {
			err = errors.Wrapf(err, "unable to unmarshel response body for %d corporation history", etagResource.ID)
			return response, etagResource, err
		}

		expires, err := RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrapf(err, "Error Encountered attempting to parse expires header for url %s", response.Path)
			return response, etagResource, err
		}
		etagResource.Expires = expires

		etag, err := RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrapf(err, "Error Encountered attempting to retrieve etag header for url %s", response.Path)
			return response, etagResource, err
		}
		etagResource.Etag = etag

		break
	case 304:
		expires, err := RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrapf(err, "Error Encountered attempting to parse expires header for url %s", response.Path)
			return response, etagResource, err
		}
		etagResource.Expires = expires

		etag, err := RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrapf(err, "Error Encountered attempting to retrieve etag header for url %s", response.Path)
			return response, etagResource, err
		}
		etagResource.Etag = etag

		break
	case 500, 502, 503, 504:
		break
	default:
		err = fmt.Errorf("Code: %d Request: %s %s", response.Code, request.Method, url.Path)
	}

	response.Data = history

	return response, etagResource, err
}

func (e *Client) PostCharactersAffiliation(ids []uint64) (Response, error) {
	var affiliations []monocle.CharacterAffiliation

	path := "/v1/characters/affiliation/"

	url := url.URL{
		Scheme: "https",
		Host:   e.Host,
		Path:   path,
	}

	headers := make(map[string]string, 0)

	// Marshal the ids []uint64 to a []byte for the request
	bIds, err := json.Marshal(ids)
	if err != nil {
		return Response{}, err
	}

	request := Request{
		Method:  "POST",
		Path:    url,
		Headers: headers,
		Body:    bIds,
	}

	response, err := e.Request(request)
	if err != nil {
		return response, err
	}

	switch response.Code {
	case 200:
		err := json.Unmarshal(response.Data.([]byte), &affiliations)
		if err != nil {
			err = errors.Wrap(err, "Unable to unmarshal response body for character affiliations")
			return response, err
		}

		break
	case 500, 502, 503, 504:
		break
	default:
		err = fmt.Errorf("Code: %d Request: %s %s", response.Code, request.Method, url.Path)
	}

	response.Data = affiliations

	return response, err
}
