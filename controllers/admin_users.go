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
)

type AdminUserController struct{}

func (c *AdminUserController) SignIn(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

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
			val.AddError("email", "Email or password was incorrect.")
			val.AddError("password", "Email or password was incorrect.")
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
		val.AddError("email", "Email or password was incorrect.")
		val.AddError("password", "Email or password was incorrect.")
		errors := map[string]interface{}{
			"errors": val.ErrorMap(),
		}
		r.JSON(res, 422, errors)
		return
	}

	// If we've reached here, email address and password were correct. Set the session and return the admin user.
	session := sessions.GetSession(req)
	session.Set("auth_token", fmt.Sprintf("%s:%s", admin.Email, admin.HashedPassword))
	r.JSON(res, 200, map[string]interface{}{"admin": admin})
}
