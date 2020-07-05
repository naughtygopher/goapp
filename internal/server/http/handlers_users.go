package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bnkamalesh/goapp/internal/users"
	"github.com/bnkamalesh/webgo/v4"
)

// CreateUser is the HTTP handler to create a new user
// This handler does not use any framework, instead just the standard library
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	u := new(users.User)
	err := json.NewDecoder(r.Body).Decode(u)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(
			[]byte(
				fmt.Sprintf("invalid request body. %s", err.Error()),
			),
		)
		return
	}

	createdUser, err := h.api.CreateUser(r.Context(), u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	b, err := json.Marshal(createdUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
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
		webgo.R500(w, err.Error())
		return
	}

	webgo.R200(w, out)
}
