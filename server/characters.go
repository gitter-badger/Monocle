package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
)

func (s *Server) handleGetCharacters(w http.ResponseWriter, r *http.Request) {

	var page uint
	var perPage uint = 1000
	urlPage := r.URL.Query().Get("page")
	if urlPage != "" {
		page64, err := strconv.ParseUint(urlPage, 10, 32)
		if err != nil {
			msg := fmt.Sprint("Bad or Invalid Page number specified")
			s.App.Logger.CriticalF("%s: %s", msg, err)
			err = errors.New(msg)
			s.WriteError(w, http.StatusBadRequest, err)
			return
		}
		page = uint(page64)
	}

	characters, err := s.App.DB.SelectCharacters(page, perPage)
	if err != nil {
		if err != sql.ErrNoRows {
			msg := fmt.Sprintf("Unable to query database for characters on page %d", page)
			s.App.Logger.CriticalF("%s: %s", msg, err)
			err = errors.New(msg)
			s.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	}

	s.WriteSuccess(w, characters, http.StatusOK)
	return
}

func (s *Server) handleGetCharacter(w http.ResponseWriter, r *http.Request) {

	urlCharacterID := chi.URLParam(r, "id")
	characterID, err := strconv.ParseUint(urlCharacterID, 10, 32)
	if err != nil {
		msg := fmt.Sprint("Unable to parse %v")
		s.App.Logger.CriticalF("%s: %s", msg, err)
		err = errors.New(msg)
		s.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	character, err := s.App.DB.SelectCharacterByCharacterID(characterID)
	if err != nil {
		if err != sql.ErrNoRows {
			msg := fmt.Sprintf("Invalid ID %d", characterID)
			s.App.Logger.CriticalF("%s: %s", msg, err)
			err = errors.New(msg)
			s.WriteError(w, http.StatusInternalServerError, err)
		}
		msg := fmt.Sprintf("Invalid ID %d", characterID)
		s.App.Logger.WarningF("%s: %s", msg, err)
		err = errors.New(msg)
		s.WriteError(w, http.StatusBadRequest, err)
		return
	}

	s.WriteSuccess(w, character, http.StatusOK)
	return
}
