package http

import (
	"encoding/json"
	"net/http"

	"github.com/naughtygopher/errors"
	"github.com/naughtygopher/goapp/internal/usernotes"
	"github.com/naughtygopher/webgo/v7"
)

func (h *Handlers) CreateUserNote(w http.ResponseWriter, r *http.Request) error {
	unote := new(usernotes.Note)
	err := json.NewDecoder(r.Body).Decode(unote)
	if err != nil {
		return errors.InputBodyErr(err, "invalid JSON provided")
	}

	un, err := h.apis.CreateUserNote(r.Context(), unote)
	if err != nil {
		return err
	}

	webgo.R200(w, un)

	return nil
}
