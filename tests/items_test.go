package tests

import (
	"bytes"
	"fmt"
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
	item := &models.Item{}
	if err := zoom.NewQuery("Item").Filter("Name =", "Test Item Create").ScanOne(item); err != nil {
		if _, ok := err.(*zoom.ModelNotFoundError); ok {
			t.Error("Item not created.")
		} else {
			panic(err)
		}
	}

	// Make sure image was actually created on s3
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
			if s3Error.StatusCode == 404 {
				t.Error("File was not created on s3.")
			} else {
				panic(err)
			}
		}
	}

	// TODO: test server-side validations
}

func TestItemsShow(t *testing.T) {
	rec := fipple.NewRecorder(t, testUrl)

	// First create a test item
	createReq := createItemRequest("Test Item Show", "An item for testing show functionality.", "0.01")
	rec.Do(createReq)

	// Get the item we created from the database
	item := &models.Item{}
	q := zoom.NewQuery("Item").Filter("Name =", "Test Item Show")
	if err := q.ScanOne(item); err != nil {
		panic(err)
	}

	// Then do a request to show it
	// NOTE: this doesn't require auth
	res := rec.Get("/items/" + item.Id)
	res.AssertOk()
	res.AssertBodyContains(`"name": "Test Item Show"`)
	res.AssertBodyContains(`"description": "An item for testing show functionality."`)
	res.AssertBodyContains(`"price": 0.01`)
}

func TestItemsUpdate(t *testing.T) {
	// TODO: fill this in
	t.Skip("Skipping update test")
}

func TestItemsDelete(t *testing.T) {
	rec := fipple.NewRecorder(t, testUrl)

	// First create a test item
	createReq := createItemRequest("Test Item Delete", "An item for testing purposes", "99.99")
	rec.Do(createReq)

	// Get the item we created from the database
	item := &models.Item{}
	if err := zoom.NewQuery("Item").Filter("Name =", "Test Item Delete").ScanOne(item); err != nil {
		panic(err)
	}

	// Create a new authenticated request for deleting the item
	deleteReq := rec.NewRequest("DELETE", "/items/"+item.Id)
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

func TestItemsIndex(t *testing.T) {
	rec := fipple.NewRecorder(t, testUrl)

	// First create two test items
	indexDescription := "This item should show up in the index."
	createReqs := []*http.Request{
		createItemRequest("Test Item Index 0", indexDescription, "99.99"),
		createItemRequest("Test Item Index 1", indexDescription, "99.99"),
	}
	for _, req := range createReqs {
		rec.Do(req)
	}

	// Get all the existing items from the database
	allItems := []*models.Item{}
	if err := zoom.NewQuery("Item").Scan(&allItems); err != nil {
		panic(err)
	}
	if len(allItems) < 2 {
		t.Errorf("Expected at least two items to be created, but got %d", len(allItems))
	}

	// Send a new index request
	// NOTE: this doesn't require authentication
	res := rec.Get("/items")
	res.AssertOk()

	// Make sure all the items we expect are there
	for _, item := range allItems {
		res.AssertBodyContains(item.Name)
		res.AssertBodyContains(item.Description)
		res.AssertBodyContains(fmt.Sprintf("%2.2f", item.Price))
		res.AssertBodyContains(item.ImageUrl)
	}
}

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
