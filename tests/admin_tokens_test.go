package tests

import (
	"github.com/albrow/fipple"
	"testing"
)

func TestAdminTokensCreate(t *testing.T) {
	rec := fipple.NewRecorder(t, testUrl)

	res := rec.Post("/admin/sign_in", map[string]string{
		"email":    "admin@5w4g.com",
		"password": "password",
	})
	res.AssertOk()
	res.AssertBodyContains("token")
}
