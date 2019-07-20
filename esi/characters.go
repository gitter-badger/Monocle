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

	// resp, err := e.Http.Do(req)
	// if err != nil {
	// 	err = errors.Wrap(err, "Unable to make request")
	// 	return response, err
	// }
	// defer resp.Body.Close()

	// if resp.StatusCode == 304 {
	// 	expires, err := time.Parse(layoutESI, resp.Header.Get("Expires"))
	// 	if err != nil {
	// 		return character, errors.Wrap(err, "Unable to parse cached_until timestamp from ESI")
	// 	}
	// 	character.Expires = expires
	// 	return character, nil
	// }
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	err = errors.Wrap(err, "Unable to read body")
	// 	return character, err
	// }

	// if resp.StatusCode >= 400 {
	// 	var apierr core.ApiError
	// 	apierr.Params = map[string]interface{}{
	// 		"id":   character.ID,
	// 		"url":  url,
	// 		"code": resp.StatusCode,
	// 	}

	// 	err = json.Unmarshal(body, &apierr)
	// 	if err != nil {
	// 		apierr.MessageErr = errors.Wrap(err, "Unable to unmarshal body")
	// 	}

	// 	return character, apierr
	// }

	// err = json.Unmarshal(body, &character)
	// if err != nil {
	// 	err = errors.Wrap(err, "Unable to unmarshal body")
	// 	return character, err
	// }

	// character.Etag = resp.Header.Get("Etag")

	// if resp.Header.Get("Expires") != "" {
	// 	expires, err := time.Parse(layoutESI, resp.Header.Get("Expires"))
	// 	if err != nil {
	// 		return character, errors.Wrap(err, "Unable to parse cached_until timestamp from ESI")
	// 	}
	// 	character.Expires = expires
	// }

	// return character, nil
}
