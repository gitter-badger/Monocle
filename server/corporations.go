package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
)

func (s *Server) handleGetCorporations(w http.ResponseWriter, r *http.Request) {

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

	corporations, err := s.App.DB.SelectCorporations(page, perPage)
	if err != nil {
		if err != sql.ErrNoRows {
			msg := fmt.Sprintf("Unable to query database for corporations on page %d", page)
			s.App.Logger.CriticalF("%s: %s", msg, err)
			err = errors.New(msg)
			s.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	}

	s.WriteSuccess(w, corporations, http.StatusOK)
	return
}

func (s *Server) handleGetCorporation(w http.ResponseWriter, r *http.Request) {

	urlCorporationID := chi.URLParam(r, "id")
	corporationID64, err := strconv.ParseUint(urlCorporationID, 10, 32)
	if err != nil {
		msg := fmt.Sprint("Unable to parse %v")
		s.App.Logger.CriticalF("%s: %s", msg, err)
		err = errors.New(msg)
		s.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	corporationID := uint(corporationID64)

	corporation, err := s.App.DB.SelectCorporationByCorporationID(corporationID)
	if err != nil {
		if err != sql.ErrNoRows {
			msg := fmt.Sprintf("Invalid ID %d", corporationID)
			s.App.Logger.CriticalF("%s: %s", msg, err)
			err = errors.New(msg)
			s.WriteError(w, http.StatusInternalServerError, err)
		}
		msg := fmt.Sprintf("Invalid ID %d", corporationID)
		s.App.Logger.WarningF("%s: %s", msg, err)
		err = errors.New(msg)
		s.WriteError(w, http.StatusBadRequest, err)
		return
	}

	s.WriteSuccess(w, corporation, http.StatusOK)
	return
}
