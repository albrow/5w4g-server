package controllers

import (
	"fmt"
	"github.com/albrow/5w4g-server/config"
	"github.com/albrow/5w4g-server/models"
	"github.com/albrow/zoom"
	"github.com/dgrijalva/jwt-go"
	"net/http"
)

func CurrentAdminUser(req *http.Request) *models.AdminUser {
	// Get the token from the request header
	token, err := jwt.ParseFromRequest(req, func(t *jwt.Token) (interface{}, error) {
		return config.PrivateKey, nil
	})
	// TODO: return 403/401 errors instead of panicking
	if err != nil {
		if err == jwt.ErrNoTokenInRequest {
			// This means the token was not provided
			fmt.Println("No token provided")
			return nil
		} else {
			// This means there was some other error
			panic(err)
		}
	}
	if token.Method != jwt.SigningMethodHS256 {
		panic(fmt.Sprintf("Incorrect signing method: %s. Expected HS256.", token.Method.Alg()))
	}

	// Get the adminId from the token claims
	adminIdInterface, found := token.Claims["adminId"]
	if !found {
		panic("No adminId provided in token.")
	}
	adminId, ok := adminIdInterface.(string)
	if !ok {
		panic("Could not convert adminId to string")
	}

	// Find the admin user with the given id in our database
	admin := &models.AdminUser{}
	if err := zoom.ScanById(adminId, admin); err != nil {
		if _, ok := err.(*zoom.KeyNotFoundError); ok {
			// This means the adminId didn't exist in the database
			panic("adminId provided in token was incorrect.")
		} else {
			// This means there was some other error
			panic(err)
		}
	}

	// TODO: check token iat against some value we store in the database for each AdminUser
	return admin
}
