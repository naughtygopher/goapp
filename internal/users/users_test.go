package users

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestUser_Sanitize(t *testing.T) {
	type fields struct {
		FirstName string
		LastName  string
		Mobile    string
		Email     string
		CreatedAt *time.Time
		UpdatedAt *time.Time
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "valid values",
			fields: fields{
				FirstName: "Jane",
				LastName:  "Doe",
				Mobile:    "9876543210",
				Email:     "jane.doe@example.com",
			},
		},
		{
			name: "with leading & trailing whitespaces",
			fields: fields{
				FirstName: "Jane ",
				LastName:  " Doe ",
				Mobile:    "  9876543210",
				Email:     "  jane.doe@example.com  ",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				FirstName: tt.fields.FirstName,
				LastName:  tt.fields.LastName,
				Mobile:    tt.fields.Mobile,
				Email:     tt.fields.Email,
				CreatedAt: tt.fields.CreatedAt,
				UpdatedAt: tt.fields.UpdatedAt,
			}
			u.Sanitize()

			trimmed := &User{
				FirstName: strings.TrimSpace(u.FirstName),
				LastName:  strings.TrimSpace(u.LastName),
				Mobile:    strings.TrimSpace(u.Mobile),
				Email:     strings.TrimSpace(u.Email),
			}

			if !reflect.DeepEqual(u, trimmed) {
				t.Fatalf("expected all trimmed values, got %v", u)
			}
		})
	}
}

func TestUser_Validate(t *testing.T) {
	type fields struct {
		FirstName string
		LastName  string
		Mobile    string
		Email     string
		CreatedAt *time.Time
		UpdatedAt *time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "all valid",
			fields: fields{
				FirstName: "Jane",
				LastName:  "Doe",
				Mobile:    "9876543210",
				Email:     "jane.doe@example.com",
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			fields: fields{
				FirstName: "Jane",
				LastName:  "Doe",
				Mobile:    "9876543210",
				Email:     "jane.doeexample.com",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				FirstName: tt.fields.FirstName,
				LastName:  tt.fields.LastName,
				Mobile:    tt.fields.Mobile,
				Email:     tt.fields.Email,
				CreatedAt: tt.fields.CreatedAt,
				UpdatedAt: tt.fields.UpdatedAt,
			}
			if err := u.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
