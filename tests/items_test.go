package tests

import (
	"bytes"
	"github.com/albrow/5w4g-server/lib"
	"github.com/albrow/5w4g-server/models"
	"github.com/albrow/fipple"
	"github.com/albrow/zoom"
	"github.com/mitchellh/goamz/s3"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
)

func createItemRequest(name string, description string, price string) *http.Request {

	// Get a valid token
	token, err := getAdminTestToken()
	if err != nil {
		panic(err)
	}

	// First, create a new multipart form writer. We need to use multipart/form-data
	// since we are including a file.
	body := bytes.NewBuffer([]byte{})
	form := multipart.NewWriter(body)
	// Add the simple key-value params to the form
	itemData := map[string]string{
		"name":        name,
		"description": description,
		"price":       price,
	}
	for fieldname, value := range itemData {
		if err := form.WriteField(fieldname, value); err != nil {
			panic(err)
		}
	}
	// Add the file to the form
	fileWriter, err := form.CreateFormFile("image", "clear.gif")
	if err != nil {
		panic(err)
	}
	// Copy the data from a test image file into the form
	testImagePath := os.Getenv("GOPATH") + "/src/github.com/albrow/5w4g-server/test_data/images/clear.gif"
	testImageFile, err := os.Open(testImagePath)
	if err != nil {
		panic(err)
	}
	if _, err := io.Copy(fileWriter, testImageFile); err != nil {
		panic(err)
	}
	// Close the form to finish writing
	if err := form.Close(); err != nil {
		panic(err)
	}
	// Create the request object and add the needed headers
	req, err := http.NewRequest("POST", testUrl+"/items", body)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "multipart/form-data; boundary="+form.Boundary())

	return req
}

func TestItemsCreate(t *testing.T) {
	rec := fipple.NewRecorder(t, testUrl)

	// Create a new authenticated request.
	req := createItemRequest("Test Item Create", "An item for testing purposes", "99.99")

	// Send the request and check the response
	res := rec.Do(req)
	res.AssertOk()
	res.AssertBodyContains("item")
	res.AssertBodyContains(`"name": "Test Item Create"`)
	res.AssertBodyContains(`"description": "An item for testing purposes"`)
	res.AssertBodyContains(`"price": 99.99`)

	// Make sure the item was actually created
	if count, err := zoom.NewQuery("Item").Filter("Name =", "Test Item Create").Count(); err != nil {
		panic(err)
	} else if count != 1 {
		t.Errorf(`Expected 1 item with name = "Test Item" to exist, but found %d items with that name.`, count)
	}

	// TODO: test server-side validations
}

func TestItemsDelete(t *testing.T) {
	rec := fipple.NewRecorder(t, testUrl)

	// Create a new authenticated request.
	createReq := createItemRequest("Test Item Delete", "An item for testing purposes", "99.99")

	// Send the request
	rec.Do(createReq)

	item := &models.Item{}
	if err := zoom.NewQuery("Item").Filter("Name =", "Test Item Delete").ScanOne(item); err != nil {
		panic(err)
	}
	deleteReq := rec.NewRequest("DELETE", "/items/"+item.Id)

	// Get a valid token and add it to the request
	token, err := getAdminTestToken()
	if err != nil {
		panic(err)
	}
	deleteReq.Header.Add("Authorization", "Bearer "+token)

	// Send the request and check the response
	res := rec.Do(deleteReq)
	res.AssertOk()

	// Make sure the item was actually deleted
	if count, err := zoom.NewQuery("Item").Filter("Name =", "Test Item Delete").Count(); err != nil {
		panic(err)
	} else if count != 0 {
		t.Error("Item was not deleted.")
	}

	// Make sure image was actually deleted from s3
	bucket, err := lib.S3Bucket()
	if err != nil {
		panic(err)
	}
	// Get the image key from the bucket
	_, err = bucket.GetKey(item.ImageOrigPath)
	if err != nil {
		// Check for an s3 error
		if s3Error, ok := err.(*s3.Error); !ok {
			panic(err)
		} else {
			if s3Error.StatusCode != 404 {
				// If the status code is 404 then the image is gone,
				// which is what we expect, if not, then there is a problem
				panic(err)
			}
		}
	} else {
		t.Error("File was not deleted from s3.")
	}
}
