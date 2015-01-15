package controllers

import (
	"code.google.com/p/go.crypto/bcrypt"
	"fmt"
	"github.com/albrow/5w4g-server/lib"
	"github.com/albrow/5w4g-server/models"
	"github.com/albrow/go-data-parser"
	"github.com/albrow/zoom"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"net/http"
)

type AdminUsersController struct{}

func (c AdminUsersController) Create(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	// Parse data from request
	adminData, err := data.Parse(req)
	if err != nil {
		panic(err)
	}

	// Validations
	val := adminData.Validator()
	val.Require("email")
	val.MatchEmail("email")
	val.Require("password")
	val.MinLength("password", 8)
	val.Require("confirmPassword")
	val.Equal("password", "confirmPassword")
	if adminData.Get("email") != "" {
		// Validate that email is unique
		count, err := zoom.NewQuery("AdminUser").Filter("Email =", adminData.Get("email")).Count()
		if err != nil {
			panic(err)
		}
		if count != 0 {
			val.AddError("email", "that email address is already taken.")
		}
	}
	if val.HasErrors() {
		r.JSON(res, 422, val.ErrorMap())
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword(adminData.GetBytes("password"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	// Save to database
	admin := &models.AdminUser{
		Email:          adminData.Get("email"),
		HashedPassword: string(hashedPassword),
	}
	if err := zoom.Save(admin); err != nil {
		panic(err)
	}

	// Render response
	r.JSON(res, 200, admin)
}

func (c AdminUsersController) Show(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	// Get the id from the url
	vars := mux.Vars(req)
	id := vars["id"]

	// Get the admin user from the database
	admin := &models.AdminUser{}
	if err := zoom.ScanById(id, admin); err != nil {
		if _, ok := err.(*zoom.KeyNotFoundError); ok {
			// This means an admin user with the given id was not found
			msg := fmt.Sprintf("Could not find admin user with id = %s", id)
			jsonErr := map[string][]string{
				"id": []string{msg},
			}
			r.JSON(res, 422, jsonErr)
			return
		} else {
			// This means there was some other error
			panic(err)
		}
	}

	// Render response
	r.JSON(res, 200, admin)
}

func (c AdminUsersController) Index(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	// Find all admin users in the database
	var admins []*models.AdminUser
	if err := zoom.NewQuery("AdminUser").Scan(&admins); err != nil {
		panic(err)
	}

	// Render response
	r.JSON(res, 200, admins)
}

func (c AdminUsersController) Delete(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	// Get the id from the url
	vars := mux.Vars(req)
	id := vars["id"]

	// Get the current user
	// Note: Earlier in the middleware chain we made sure currentUser was not nil
	currentUser := lib.CurrentAdminUser(req)

	// Sanity check. (You can't delete yourself)
	if currentUser.Id == id {
		// This means an admin user with the given id was not found
		jsonErr := map[string][]string{
			"id": []string{"You can't delete yourself, dummy!"},
		}
		r.JSON(res, 422, jsonErr)
		return
	}

	// Delete from database
	if err := zoom.DeleteById("AdminUser", id); err != nil {
		panic(err)
	}

	// Render response
	r.JSON(res, 200, struct{}{})
}
