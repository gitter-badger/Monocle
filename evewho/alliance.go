package evewho

import (
	"net/url"
	"strconv"
)

type (
	AllianceList struct {
		Info       AlliShortInfo `json:"info"`
		Characters []AlliChar    `json:"characters"`
	}
	AlliShortInfo struct {
		AllianceID  string `json:"alliance_id"`
		Name        string `json:"name"`
		MemberCount string `json:"memberCount"`
	}
	AlliChar struct {
		CharacterID   string `json:"character_id"`
		CorporationID string `json:"corporation_id"`
		AllianceID    string `json:"alliance_id"`
		Name          string `json:"name"`
	}
)

func (e *Client) GetAllianceMembersByID(id uint, page int) (Response, error) {

	v := url.Values{}
	v.Set("type", "allilist")
	v.Set("id", strconv.FormatUint(uint64(id), 10))
	if page > 0 {
		v.Set("page", strconv.FormatUint(uint64(page), 10))
	}
	query := v.Encode()

	uri := url.URL{
		Scheme:   "https",
		Host:     e.Host,
		Path:     "api.php",
		RawQuery: query,
	}

	headers := make(map[string]string)

	request := Request{
		Method:  "GET",
		Path:    uri,
		Headers: headers,
		Body:    nil,
	}

	return e.Request(request)

}
