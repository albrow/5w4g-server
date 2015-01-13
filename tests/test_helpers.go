package tests

import (
	"bytes"
	"fmt"
	"github.com/OneOfOne/xxhash/native"
	"github.com/albrow/5w4g-server/config"
	"github.com/albrow/5w4g-server/lib"
	"github.com/albrow/5w4g-server/models"
	"github.com/albrow/zoom"
	"github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/goamz/s3"
	"io"
	"os"
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
	// whether or not config has been initialized
	configIsInit = false
)

func getAdminTestToken() (string, error) {
	if !configIsInit {
		config.Env = "test"
		config.Init()
		configIsInit = true
	}

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
	if token, err := token.SignedString(config.PrivateKey); err != nil {
		return token, err
	} else {
		adminTestToken = token
		return token, nil
	}
}

func getAdminTestUser() (*models.AdminUser, error) {
	if !configIsInit {
		config.Env = "test"
		config.Init()
		configIsInit = true
	}

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

// s3FileExists returns true iff there is a file on the s3 bucket designated
// by path. It panics if there are any errors connecting to the bucket.
func s3FileExists(path string) bool {
	// Get the bucket
	bucket, err := lib.S3Bucket()
	if err != nil {
		panic(err)
	}
	// Attmpt to get the key from the bucket
	_, err = bucket.GetKey(path)
	if err != nil {
		// Check for an s3 error
		if s3Error, ok := err.(*s3.Error); !ok {
			panic(err)
		} else {
			if s3Error.StatusCode == 404 {
				return false
			} else {
				panic(err)
			}
		}
	}
	return true
}

// calculateHashForFile calculates a hash for the file at the given path.
// It panics if there were any errrors opening the file or calculating the hash.
func calculateHashForFile(path string) string {
	h := xxhash.New64()
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	io.Copy(h, f)
	return string(h.Sum(nil))
}

// calculateHashForPath calculates a hash for the file stored on s3 designated
// by path. It panics if there were any errrors communicated with s3, opening the file
// or calculating the hash.
func calculateHashForS3File(path string) string {
	bucket, err := lib.S3Bucket()
	if err != nil {
		panic(err)
	}
	contents, err := bucket.Get(path)
	if err != nil {
		panic(err)
	}
	h := xxhash.New64()
	io.Copy(h, bytes.NewBuffer(contents))
	return string(h.Sum(nil))
}
