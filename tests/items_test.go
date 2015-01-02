package tests

import (
	"bytes"
	"github.com/albrow/fipple"
	"github.com/albrow/zoom"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
)

func TestItemsCreate(t *testing.T) {
	rec := fipple.NewRecorder(t, testUrl)

	// Get a valid token
	token, err := getAdminTestToken()
	if err != nil {
		panic(err)
	}

	// Create a new authenticated request. We need to use multipart/form-data
	// since we are including a file.
	// First, create a new multipart form writer
	body := bytes.NewBuffer([]byte{})
	form := multipart.NewWriter(body)
	// Add the simple key-value params to the form
	itemData := map[string]string{
		"name":        "Test Item",
		"description": "An item for testing purposes",
		"price":       "99.99",
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

	// Send the request and check the response
	res := rec.Do(req)
	res.AssertOk()
	res.AssertBodyContains("item")
	res.AssertBodyContains(`"name": "Test Item"`)
	res.AssertBodyContains(`"description": "An item for testing purposes"`)
	res.AssertBodyContains(`"price": 99.99`)

	// Make sure the user was actually created
	if count, err := zoom.NewQuery("Item").Filter("Name =", "Test Item").Count(); err != nil {
		panic(err)
	} else if count != 1 {
		t.Errorf(`Expected 1 item with name = "Test Item" to exist, but found %d items with that name.`, count)
	}

	// TODO: test server-side validations
}
