package tests

import (
	"github.com/albrow/fipple"
	"testing"
)

func TestAdminTokensCreate(t *testing.T) {
	rec := fipple.NewRecorder(t, testUrl)

	// Get the admin user
	admin, err := getAdminTestUser()
	if err != nil {
		panic(err)
	}

	// First sign in with a valid email and password
	res := rec.Post("/admin_users/sign_in", map[string]string{
		"email": admin.Email,
		// we have no choice but to assume the default admin password
		"password": "password",
	})
	res.AssertOk()
	res.AssertBodyContains("token")

	// Use a table-driven test to check for proper error responses
	// TODO: can we generalize the test input concept for other tests?
	testInputs := []struct {
		email            string
		password         string
		expectedContains []string
		expectedCode     int
	}{
		{
			// Correct email, wrong password
			email:    admin.Email,
			password: "not_the_real_password",
			expectedContains: []string{
				"error",
				"email or password was incorrect",
			},
			expectedCode: 422,
		},
		{
			// Correct password, wrong email
			email:    "not_the_real_email@foo.com",
			password: "password",
			expectedContains: []string{
				"error",
				"email or password was incorrect",
			},
			expectedCode: 422,
		},
	}

	for _, testInput := range testInputs {
		res := rec.Post("/admin_users/sign_in", map[string]string{
			"email":    testInput.email,
			"password": testInput.password,
		})
		res.AssertCode(testInput.expectedCode)
		for _, txt := range testInput.expectedContains {
			res.AssertBodyContains(txt)
		}
	}
}
