package tests

import (
	"github.com/albrow/fipple"
	"testing"
)

func TestAdminUsersCreate(t *testing.T) {
	rec := fipple.NewRecorder(t, testUrl)

	// Get a valid token
	token, err := getAdminTestToken()
	if err != nil {
		panic(err)
	}

	// Created an authenticated request
	req := rec.NewRequestWithData("POST", "/admin/users", map[string]string{
		"email":           "test@example.com",
		"password":        "password",
		"confirmPassword": "password",
	})
	req.Header.Add("Authorization", "Bearer "+token)

	// Send the request and check the response
	res := rec.Do(req)
	res.AssertOk()
	res.AssertBodyContains("admin")
	res.AssertBodyContains(`"email": "test@example.com"`)
}
