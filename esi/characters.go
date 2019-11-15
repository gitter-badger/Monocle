package esi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/pkg/errors"
)

// HeadCharactersCharacterID makes a HTTP GET Request to the /characters/{character_id} endpoint
// Often used to see if a particular character exists or to check the remaining time until
// the cache expires
//
// Documentation: https://esi.evetech.net/ui/#/Character/get_characters_character_id
// Version: v4
// Cache: 3600 sec (1 Hour)
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

// GetCharactersCharacterID makes a HTTP GET Request to the /characters/{character_id} endpoint
// for information about the provided character
//
// Documentation: https://esi.evetech.net/ui/#/Character/get_characters_character_id
// Version: v4
// Cache: 3600 sec (1 Hour)
func (e *Client) GetCharactersCharacterID(character *monocle.Character) (Response, error) {

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

		newChar.Expires, err = RetrieveExpiresHeaderFromResponse(response, 0)
		if err != nil {
			return response, errors.Wrap(err, "Error Encountered attempting to parse expires header")
		}

		newChar.Etag, err = RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			return response, errors.Wrap(err, "Error Encountered attempting to retrieve etag header")
		}

		character = &newChar

		break
	case 304:
		character.Expires, err = RetrieveExpiresHeaderFromResponse(response, 0)
		if err != nil {
			return response, errors.Wrap(err, "Error Encountered attempting to parse expires header")
		}

		character.Etag, err = RetrieveEtagHeaderFromResponse(response)
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

	response.Data = character

	return response, err
}

// GetCharactersCharacterIDCorporationHistory makes a HTTP GET Request to the
// /characters/{character_id}/corporationhistory endpoint for a list of corporations
// the character has previously been a member of
//
// Documentation: https://esi.evetech.net/ui/#/Character/get_characters_character_id_corporationhistory
// Version: v1
// Cache: 3600 sec (1 Hour)
func (e *Client) GetCharactersCharacterIDCorporationHistory(etag *monocle.EtagResource) (Response, error) {

	var history []*monocle.CharacterCorporationHistory

	path := fmt.Sprintf("/v1/characters/%d/corporationhistory/", etag.ID)

	url := url.URL{
		Scheme: "https",
		Host:   e.Host,
		Path:   path,
	}

	headers := make(map[string]string)

	if etag.Etag != "" {
		headers["If-None-Match"] = etag.Etag
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
			return response, errors.Wrapf(err, "unable to unmarshel response body for %d corporation history", etag.ID)
		}

		etag.Expires, err = RetrieveExpiresHeaderFromResponse(response, 0)
		if err != nil {
			return response, errors.Wrapf(err, "Error Encountered attempting to parse expires header for url %s", response.Path)
		}

		etag.Etag, err = RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			return response, errors.Wrapf(err, "Error Encountered attempting to retrieve etag header for url %s", response.Path)
		}

		break
	case 304:
		etag.Expires, err = RetrieveExpiresHeaderFromResponse(response, 0)
		if err != nil {
			return response, errors.Wrapf(err, "Error Encountered attempting to parse expires header for url %s", response.Path)
		}

		etag.Etag, err = RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			return response, errors.Wrapf(err, "Error Encountered attempting to retrieve etag header for url %s", response.Path)
		}

		break
	case 500, 502, 503, 504:
		break
	default:
		err = fmt.Errorf("Code: %d Request: %s %s", response.Code, request.Method, url.Path)
	}

	response.Data = map[string]interface{}{
		"etag":    etag,
		"history": history,
	}

	return response, err
}

// PostCharactersAffiliation makes a HTTP POST Request to the
// /characters/affiliation/ endpoint containing up to 1K character ids
// This is often used to quickly determine if a characters affiliation with an alliance,
// corporation, or faction has recently changed.
//
// Documentation: https://esi.evetech.net/ui/#/Character/post_characters_affiliation
// Version: v1
// Cache: 3600 sec (1 Hour)
func (e *Client) PostCharactersAffiliation(ids []uint64) (Response, error) {
	var affiliations []monocle.CharacterAffiliation

	path := "/v1/characters/affiliation/"

	url := url.URL{
		Scheme: "https",
		Host:   e.Host,
		Path:   path,
	}

	headers := make(map[string]string)

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
			return response, errors.Wrap(err, "Unable to unmarshal response body for character affiliations")
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
