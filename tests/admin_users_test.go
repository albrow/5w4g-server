package tests

import (
	"fmt"
	"github.com/albrow/5w4g-server/models"
	"github.com/albrow/fipple"
	"github.com/albrow/zoom"
	"testing"
)

func TestAdminUsersCreate(t *testing.T) {
	rec := fipple.NewRecorder(t, testUrl)

	// Get a valid token
	token, err := getAdminTestToken()
	if err != nil {
		panic(err)
	}

	// Create an authenticated request
	req := rec.NewRequestWithData("POST", "/admin_users", map[string]string{
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

	// Make sure the user was actually created
	if count, err := zoom.NewQuery("AdminUser").Filter("Email =", "test@example.com").Count(); err != nil {
		panic(err)
	} else if count != 1 {
		t.Errorf("Expected 1 admin user with email %s to exist, but found %d admin users with that email.", "test@example.com", count)
	}
}

func TestAdminUsersShow(t *testing.T) {
	rec := fipple.NewRecorder(t, testUrl)

	// Get a valid token
	token, err := getAdminTestToken()
	if err != nil {
		panic(err)
	}

	// Get the id of the current admin user
	admin, err := getAdminTestUser()
	if err != nil {
		panic(err)
	}

	// Create an authenticated request
	req := rec.NewRequest("GET", "/admin_users/"+admin.Id)
	req.Header.Add("Authorization", "Bearer "+token)

	// Send the request and check the response
	res := rec.Do(req)
	res.AssertOk()
	res.AssertBodyContains("admin")
	res.AssertBodyContains(`"email": "admin@5w4g.com"`)
	res.AssertBodyContains(fmt.Sprintf(`"id": "%s"`, admin.Id))
}

func TestAdminUsersIndex(t *testing.T) {
	rec := fipple.NewRecorder(t, testUrl)

	// Get a valid token
	token, err := getAdminTestToken()
	if err != nil {
		panic(err)
	}

	// Create an authenticated request
	req := rec.NewRequest("GET", "/admin_users")
	req.Header.Add("Authorization", "Bearer "+token)

	// Send the request and check the response
	res := rec.Do(req)
	res.AssertOk()
	res.AssertBodyContains("admins")
	res.AssertBodyContains(`"email": "admin@5w4g.com"`)
}

func TestAdminUsersDelete(t *testing.T) {
	rec := fipple.NewRecorder(t, testUrl)

	// Get a valid token
	token, err := getAdminTestToken()
	if err != nil {
		panic(err)
	}

	// Create a new admin user (which we will then delete)
	createReq := rec.NewRequestWithData("POST", "/admin_users", map[string]string{
		"email":           "delete@me.com",
		"password":        "password",
		"confirmPassword": "password",
	})
	createReq.Header.Add("Authorization", "Bearer "+token)
	createRes := rec.Do(createReq)
	createRes.AssertOk()

	// Get the id of the user we just created
	admin := &models.AdminUser{}
	if err := zoom.NewQuery("AdminUser").Filter("Email =", "delete@me.com").ScanOne(admin); err != nil {
		panic(err)
	}

	// Create an authenticated request to delete the admin user
	deleteReq := rec.NewRequest("DELETE", "/admin_users/"+admin.Id)
	deleteReq.Header.Add("Authorization", "Bearer "+token)

	// Send the request and check the response
	res := rec.Do(deleteReq)
	res.AssertOk()

	// Make sure the user was actually deleted
	if count, err := zoom.NewQuery("AdminUser").Filter("Email =", "delete@me.com").Count(); err != nil {
		panic(err)
	} else if count != 0 {
		t.Errorf("Expected admin user with email %s to be deleted, but found %d admin users with that email.", admin.Email, count)
	}
}
