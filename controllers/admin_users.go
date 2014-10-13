package controllers

import (
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/albrow/5w4g-server/models"
	"github.com/albrow/go-data-parser"
	"github.com/albrow/zoom"
	"github.com/unrolled/render"
	"net/http"
)

type AdminUsersController struct{}

func (c AdminUsersController) Create(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	// Make sure we're signed in
	if currentUser := CurrentAdminUser(req); currentUser == nil {
		jsonData := map[string]interface{}{
			"errors": map[string][]string{
				"error": []string{"You need to be signed in to do that!"},
			},
		}
		r.JSON(res, 401, jsonData)
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
		errors := map[string]interface{}{
			"errors": val.ErrorMap(),
		}
		r.JSON(res, 422, errors)
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
	jsonData := map[string]interface{}{
		"admin":   admin,
		"message": "New admin user created!",
	}
	r.JSON(res, 200, jsonData)
}

func (c AdminUsersController) Index(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	// Make sure we're signed in
	if currentUser := CurrentAdminUser(req); currentUser == nil {
		jsonData := map[string]interface{}{
			"errors": map[string][]string{
				"error": []string{"You need to be signed in to do that!"},
			},
		}
		r.JSON(res, 401, jsonData)
		return
	}

	// Find all admin users in the database
	var admins []*models.AdminUser
	if err := zoom.NewQuery("AdminUser").Scan(&admins); err != nil {
		panic(err)
	}

	// render response
	dataMap := map[string]interface{}{"admins": admins}
	r.JSON(res, 200, dataMap)
}
