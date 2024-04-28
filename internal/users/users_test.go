package users

import (
	"reflect"
	"testing"
)

func TestUser_Sanitize(t *testing.T) {
	tests := []struct {
		name   string
		input  User
		output User
	}{
		{
			name: "sanitize",
			input: User{
				ID:             " ID ",
				FullName:       " Fullname ",
				Email:          " Email ",
				Phone:          " Phone ",
				ContactAddress: " Contact Address ",
			},
			output: User{
				ID:             "ID",
				FullName:       "Fullname",
				Email:          "Email",
				Phone:          "Phone",
				ContactAddress: "Contact Address",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.Sanitize()
			if !reflect.DeepEqual(tt.input, tt.output) {
				t.Errorf("got: %+v, expected: %+v", tt.input, tt.output)
			}
		})
	}
}

func TestUser_ValidateForCreate(t *testing.T) {
	type fields struct {
		ID             string
		FullName       string
		Email          string
		Phone          string
		ContactAddress string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "no error",
			fields: fields{
				ID:             "",
				FullName:       "Full Name",
				Email:          "name@example.com",
				Phone:          "+91 1234567890",
				ContactAddress: "Addr line 1, line 2, City, PIN, Country",
			},
			wantErr: false,
		},
		{
			name: "no name",
			fields: fields{
				ID:             "ID::1",
				FullName:       "",
				Email:          "name@example.com",
				Phone:          "+91 1234567890",
				ContactAddress: "Addr line 1, line 2, City, PIN, Country",
			},
			wantErr: true,
		},
		{
			name: "no email",
			fields: fields{
				ID:             "ID::1",
				FullName:       "Full Name",
				Email:          "",
				Phone:          "+91 1234567890",
				ContactAddress: "Addr line 1, line 2, City, PIN, Country",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			us := &User{
				ID:             tt.fields.ID,
				FullName:       tt.fields.FullName,
				Email:          tt.fields.Email,
				Phone:          tt.fields.Phone,
				ContactAddress: tt.fields.ContactAddress,
			}
			if err := us.ValidateForCreate(); (err != nil) != tt.wantErr {
				t.Errorf("User.ValidateForCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
