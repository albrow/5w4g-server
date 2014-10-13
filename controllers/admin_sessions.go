package controllers

import (
	"code.google.com/p/go.crypto/bcrypt"
	"fmt"
	"github.com/albrow/5w4g-server/models"
	"github.com/albrow/go-data-parser"
	"github.com/albrow/zoom"
	"github.com/goincremental/negroni-sessions"
	"github.com/unrolled/render"
	"net/http"
	"strings"
)

type AdminSessionsController struct{}

func (c *AdminSessionsController) Create(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	// Check if admin is already signed in
	if admin := CurrentAdminUser(req); admin != nil {
		r.JSON(res, 200, map[string]interface{}{
			"admin":           admin,
			"message":         "You were already signed in!",
			"alreadySignedIn": true,
		})
		return
	}

	// Parse request body
	adminData, err := data.Parse(req)
	if err != nil {
		panic(err)
	}

	// Validate
	val := adminData.Validator()
	val.Require("email")
	val.MatchEmail("email")
	val.Require("password")
	if val.HasErrors() {
		errors := map[string]interface{}{
			"errors": val.ErrorMap(),
		}
		r.JSON(res, 422, errors)
		return
	}

	// Find the admin by email address
	admin := &models.AdminUser{}
	if err := zoom.NewQuery("AdminUser").Filter("Email =", adminData.Get("email")).ScanOne(admin); err != nil {
		if _, ok := err.(*zoom.ModelNotFoundError); ok {
			// This means a model with that email address was not found
			val.AddError("email", "email or password was incorrect.")
			errors := map[string]interface{}{
				"errors": val.ErrorMap(),
			}
			r.JSON(res, 422, errors)
			return
		} else {
			// This means there was an error connecting to the database
			panic(err)
		}
	}

	// Check if the found admin's password matches the submitted password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.HashedPassword), adminData.GetBytes("password")); err != nil {
		val.AddError("email", "email or password was incorrect.")
		errors := map[string]interface{}{
			"errors": val.ErrorMap(),
		}
		r.JSON(res, 422, errors)
		return
	}

	// If we've reached here, email address and password were correct. Set the session and return the admin user.
	session := sessions.GetSession(req)
	session.Set("auth_token", fmt.Sprintf("%s:%s", admin.Email, admin.HashedPassword))
	r.JSON(res, 200, map[string]interface{}{
		"admin":           admin,
		"message":         "You are now signed in.",
		"alreadySignedIn": false,
	})
}

func (c *AdminSessionsController) Delete(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	// Check if admin is already signed in
	if admin := CurrentAdminUser(req); admin == nil {
		r.JSON(res, 200, map[string]interface{}{
			"message":          "You were already signed out!",
			"alreadySignedOut": true,
		})
		return
	}

	// Delete auth_token from the session data
	session := sessions.GetSession(req)
	session.Delete("auth_token")
	r.JSON(res, 200, map[string]interface{}{
		"message":          "You have been signed out.",
		"alreadySignedOut": false,
	})
}

func (c *AdminSessionsController) Show(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	// Check if admin is signed in
	admin := CurrentAdminUser(req)
	if admin != nil {
		r.JSON(res, 200, map[string]interface{}{
			"admin":    admin,
			"message":  "You are signed in.",
			"signedIn": true,
		})
		return
	} else {
		r.JSON(res, 200, map[string]interface{}{
			"message":  "You are not signed in.",
			"signedIn": false,
		})
		return
	}
}

func CurrentAdminUser(req *http.Request) *models.AdminUser {
	// Get and parse the auth_token from the session data
	session := sessions.GetSession(req)
	token := session.Get("auth_token")
	if token == nil {
		return nil
	}
	tokenString, ok := token.(string)
	if !ok {
		panic("Could not convert auth_token to string")
	}
	split := strings.SplitN(tokenString, ":", 2)
	email, hashedPassword := split[0], split[1]

	// Find the admin user with the provided email and password
	admin := &models.AdminUser{}
	q := zoom.NewQuery("AdminUser").Filter("Email =", email).Filter("HashedPassword =", hashedPassword)
	if err := q.ScanOne(admin); err != nil {
		if _, ok := err.(*zoom.ModelNotFoundError); ok {
			// This means an admin user was not found with the given email and password.
			// We return nil to indicate that there is not a user currently logged in.
			return nil
		} else {
			// There was an error connecting to the database.
			panic(err)
		}
	}
	return admin
}
