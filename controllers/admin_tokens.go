package controllers

import (
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/albrow/5w4g-server/config"
	"github.com/albrow/5w4g-server/models"
	"github.com/albrow/go-data-parser"
	"github.com/albrow/zoom"
	"github.com/dgrijalva/jwt-go"
	"github.com/unrolled/render"
	"net/http"
	"time"
)

type AdminTokensController struct{}

func (c *AdminTokensController) Create(res http.ResponseWriter, req *http.Request) {
	r := render.New()

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
		r.JSON(res, 422, val.ErrorMap())
		return
	}

	// Find the admin by email address
	admin := &models.AdminUser{}
	if err := zoom.NewQuery("AdminUser").Filter("Email =", adminData.Get("email")).ScanOne(admin); err != nil {
		if _, ok := err.(*zoom.ModelNotFoundError); ok {
			// This means a model with that email address was not found
			val.AddError("email", "email or password was incorrect.")
			r.JSON(res, 422, val.ErrorMap())
			return
		} else {
			// This means there was an error connecting to the database
			panic(err)
		}
	}

	// Check if the found admin's password matches the submitted password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.HashedPassword), adminData.GetBytes("password")); err != nil {
		val.AddError("email", "email or password was incorrect.")
		r.JSON(res, 422, val.ErrorMap())
		return
	}

	// If we've reached here, email address and password were correct.
	// Create and return a signed JWT
	token := jwt.New(jwt.SigningMethodHS256)
	// Store some claims in the token
	token.Claims["adminId"] = admin.Id
	now := time.Now().UTC()
	// Expires 30 days from now. Formatted as unix time in UTC
	token.Claims["exp"] = now.Add(24 * time.Hour * 30).Unix()
	// iat is the time the token was created. We can use this to revoke tokens
	// created before a certain time (the time an account was compromised).
	token.Claims["iat"] = now.Unix()

	// Sign the token with our private key
	signedToken, err := token.SignedString(config.PrivateKey)
	if err != nil {
		panic(err)
	}

	r.JSON(res, http.StatusOK, map[string]interface{}{
		"token": signedToken,
	})
}
