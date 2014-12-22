package tests

import (
	"fmt"
	"github.com/albrow/5w4g-server/config"
	"github.com/albrow/5w4g-server/models"
	"github.com/albrow/zoom"
	"github.com/dgrijalva/jwt-go"
	"time"
)

var (
	// testUrl is the url to use for all integration tests.
	testUrl = fmt.Sprintf("http://%s:%s", config.Test.Host, config.Test.Port)
	// adminTestToken is a token with claims matching a valid admin
	// user, which can be used for testing. It is created as needed
	// in the getAdminTestToken function. A single adminTestToken is
	// used for all tests every time tests are run and a new one
	// is generated every time tests are run.
	adminTestToken = ""
	// adminTestUser is an AdminUser that can be used for testing.
	// It is created as needed with the getAdminTestUser function,
	// stored in the test database, and reused for successive tests.
	adminTestUser *models.AdminUser = nil
)

func getAdminTestToken() (string, error) {
	// If adminTestToken was previously created, return it
	if adminTestToken != "" {
		return adminTestToken, nil
	}

	// Otherwise we will need to create a new token
	token := jwt.New(jwt.SigningMethodHS256)

	// Store some claims in the token associated with adminTestUser
	admin, err := getAdminTestUser()
	if err != nil {
		return "", err
	}
	token.Claims["adminId"] = admin.Id
	now := time.Now().UTC()
	token.Claims["exp"] = now.Add(24 * time.Hour * 30).Unix()
	token.Claims["iat"] = now.Unix()

	// Sign the token with our testing private key.
	// Since this is strictly for testing purposes, it is okay
	// to use a non-random, non-private key here.
	if token, err := token.SignedString("TEST_PRIVATE_KEY"); err != nil {
		return token, err
	} else {
		adminTestToken = token
		return token, nil
	}
}

func getAdminTestUser() (*models.AdminUser, error) {
	// If adminTestUser was previously created, return it
	if adminTestUser != nil {
		return adminTestUser, nil
	}

	// Otherwise get the user from the database
	models.Init()
	admin := &models.AdminUser{}
	if err := zoom.NewQuery("AdminUser").Filter("Email =", "admin@5w4g.com").ScanOne(admin); err != nil {
		if _, ok := err.(*zoom.ModelNotFoundError); ok {
			// If the expected admin user doesn't exist, it's a problem.
			// We don't really expect this to happen since the default
			// admin user is created on server start if it doesn't exist.
			// Hoever, this check is here just in case we eff something up
			// and the user doesn't exist.
			return nil, fmt.Errorf("The default admin user did not exist. Cannot continue with test.")
		} else {
			// If there was some other error, return it
			return nil, err
		}
	}
	adminTestUser = admin
	return admin, nil
}
