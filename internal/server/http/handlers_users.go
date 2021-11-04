package http

import (
	"encoding/json"
	"net/http"

	"github.com/bnkamalesh/errors"
	"github.com/bnkamalesh/goapp/internal/users"
	"github.com/bnkamalesh/webgo/v6"
)

// CreateUser is the HTTP handler to create a new user
// This handler does not use any framework, instead just the standard library
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	u := new(users.User)
	err := json.NewDecoder(r.Body).Decode(u)
	if err != nil {
		errResponder(w, errors.InputBodyErr(err, "invalid JSON provided"))
		return
	}

	createdUser, err := h.api.CreateUser(r.Context(), u)
	if err != nil {
		errResponder(w, err)
		return
	}

	b, err := json.Marshal(createdUser)
	if err != nil {
		errResponder(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

// ReadUserByEmail is the HTTP handler to read an existing user by email
func (h *Handlers) ReadUserByEmail(w http.ResponseWriter, r *http.Request) {
	wctx := webgo.Context(r)
	email := wctx.Params()["email"]

	out, err := h.api.ReadUserByEmail(r.Context(), email)
	if err != nil {
		errResponder(w, err)
		return
	}

	webgo.R200(w, out)
}
