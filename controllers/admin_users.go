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

	// Make sure we're signed in
	if currentUser := lib.CurrentAdminUser(req); currentUser == nil {
		r.JSON(res, 401, lib.ErrUnauthorized)
		return
	}

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

	// Make sure we're signed in
	if currentUser := lib.CurrentAdminUser(req); currentUser == nil {
		r.JSON(res, 401, lib.ErrUnauthorized)
		return
	}

	// Get the id from the url
	vars := mux.Vars(req)
	id, found := vars["id"]
	if !found {
		r.JSON(res, 422, lib.NewJsonError("Missing required url parameter: id"))
		return
	}

	// Get the admin user from the database
	admin := &models.AdminUser{}
	if err := zoom.ScanById(id, admin); err != nil {
		if _, ok := err.(*zoom.KeyNotFoundError); ok {
			// This means an admin user with the given id was not found
			msg := fmt.Sprintf("Could not find admin user with id = %s", id)
			r.JSON(res, 422, lib.NewJsonError(msg))
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

	// Make sure we're signed in
	if currentUser := lib.CurrentAdminUser(req); currentUser == nil {
		r.JSON(res, 401, lib.ErrUnauthorized)
		return
	}

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

	// Make sure we're signed in
	currentUser := lib.CurrentAdminUser(req)
	if currentUser == nil {
		r.JSON(res, 401, lib.ErrUnauthorized)
		return
	}

	// Get the id from the url
	vars := mux.Vars(req)
	id, found := vars["id"]
	if !found {
		r.JSON(res, 422, lib.NewJsonError("Missing required url parameter: id"))
		return
	}

	// Sanity check. (You can't delete yourself)
	if currentUser.Id == id {
		r.JSON(res, 422, lib.NewJsonError("You can't delete yourself, dummy!"))
		return
	}

	// Delete from database
	if err := zoom.DeleteById("AdminUser", id); err != nil {
		panic(err)
	}

	// Render response
	r.JSON(res, 200, struct{}{})
}
