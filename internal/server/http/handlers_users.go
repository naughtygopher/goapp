package http

import (
	"encoding/json"
	"net/http"

	"github.com/bnkamalesh/errors"
	"github.com/bnkamalesh/webgo/v6"

	"github.com/bnkamalesh/goapp/internal/users"
)

// CreateUser is the HTTP handler to create a new user
// This handler does not use any framework, instead just the standard library
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) error {
	u := new(users.User)
	err := json.NewDecoder(r.Body).Decode(u)
	if err != nil {
		return errors.InputBodyErr(err, "invalid JSON provided")
	}

	createdUser, err := h.api.CreateUser(r.Context(), u)
	if err != nil {
		return err
	}

	b, err := json.Marshal(createdUser)
	if err != nil {
		return errors.InputBodyErr(err, "invalid input body provided")
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(b)
	if err != nil {
		return errors.Wrap(err, "failed to respond")
	}
	return nil
}

// ReadUserByEmail is the HTTP handler to read an existing user by email
func (h *Handlers) ReadUserByEmail(w http.ResponseWriter, r *http.Request) error {
	wctx := webgo.Context(r)
	email := wctx.Params()["email"]

	out, err := h.api.ReadUserByEmail(r.Context(), email)
	if err != nil {
		return err
	}

	webgo.R200(w, out)
	return nil
}
