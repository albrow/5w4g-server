package tests

import (
	"fmt"
	"github.com/albrow/5w4g-server/config"
	"github.com/albrow/5w4g-server/models"
	"github.com/albrow/fipple"
	"github.com/albrow/zoom"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

var (
	blueImage      = filepath.Join(config.AppRoot, "test_data", "images", "blue.gif")
	redImage       = filepath.Join(config.AppRoot, "test_data", "images", "red.gif")
	clearImage     = filepath.Join(config.AppRoot, "test_data", "images", "clear.gif")
	blueImageHash  = calculateHashForFile(blueImage)
	redImageHash   = calculateHashForFile(redImage)
	clearImageHash = calculateHashForFile(clearImage)
)

func TestItemsCreate(t *testing.T) {
	t.Parallel()
	rec := fipple.NewRecorder(t, testUrl)

	// Create a new authenticated request.
	createDesc := "An item for testing purposes"
	createName := "Test Item Create"
	createPrice := "99.99"
	req := createItemRequest(rec, map[string]string{
		"name":        createName,
		"description": createDesc,
		"price":       createPrice,
	}, blueImage)

	// Send the request and check the response
	res := rec.Do(req)
	res.AssertOk()
	res.AssertBodyContains("item")
	res.AssertBodyContains(fmt.Sprintf(`"name": "%s"`, createName))
	res.AssertBodyContains(fmt.Sprintf(`"description": "%s"`, createDesc))
	res.AssertBodyContains(`"price": ` + createPrice)

	// Make sure the item was actually created
	item := &models.Item{}
	if err := zoom.NewQuery("Item").Filter("Name =", createName).ScanOne(item); err != nil {
		if _, ok := err.(*zoom.ModelNotFoundError); ok {
			t.Error("Item not created.")
		} else {
			panic(err)
		}
	}

	// Make sure image was actually created on s3 and that its contents match what we expect
	if !s3FileExists(item.ImageS3Path) {
		t.Error("File was not created on s3.")
	}
	if calculateHashForS3File(item.ImageS3Path) != blueImageHash {
		t.Error("The s3 file hash did not equal what we expected it to be (the blue image hash). Therefore, image file was not actually uploaded.")
	}

	// Attempting to create a second item with the same name should fail
	req = createItemRequest(rec, map[string]string{
		"name":        createName,
		"description": createDesc,
		"price":       createPrice,
	}, blueImage)
	res = rec.Do(req)
	res.AssertCode(422)
	res.AssertBodyContains(`"name"`)
	res.AssertBodyContains("already taken")
}

func TestItemsShow(t *testing.T) {
	rec := fipple.NewRecorder(t, testUrl)

	// First create a test item
	showName := "Test Item Show"
	showDesc := "An item for testing show functionality."
	showPrice := 0.99
	item := createMockItem(showName, showDesc, showPrice)

	// Then do a request to show it
	// NOTE: this doesn't require auth
	res := rec.Get("/items/" + item.Id)
	res.AssertOk()
	res.AssertBodyContains(fmt.Sprintf(`"name": "%s"`, showName))
	res.AssertBodyContains(fmt.Sprintf(`"description": "%s"`, showDesc))
	res.AssertBodyContains(fmt.Sprintf(`"price": %1.2f`, showPrice))
}

func TestItemsUpdate(t *testing.T) {
	t.Parallel()
	rec := fipple.NewRecorder(t, testUrl)

	// First create a test item
	origName := "Test Item Update 0"
	origDesc := "An item for testing update functionality."
	origPrice := "0.99"
	createReq := createItemRequest(rec, map[string]string{
		"name":        origName,
		"description": origDesc,
		"price":       origPrice,
	}, blueImage)
	rec.Do(createReq)

	// Get the item we created from the database
	origItem := &models.Item{}
	q := zoom.NewQuery("Item").Filter("Name =", origName)
	if err := q.ScanOne(origItem); err != nil {
		panic(err)
	}

	// Now we want to test a few different update cases and make sure
	// everything works on the backend. Especially with the somewhat
	// complicated interaction between our server and S3.
	// --
	// CASE 0: First try just updating the description and price.
	newDesc := "An updated description"
	newPrice := "1.05"
	updateDescPriceReq := updateItemRequest(rec, origItem.Id, map[string]string{
		"description": newDesc,
		"price":       newPrice,
	}, "")
	res := rec.Do(updateDescPriceReq)

	// Check the response
	res.AssertOk()
	res.AssertBodyContains(fmt.Sprintf(`"name": "%s"`, origName))
	res.AssertBodyContains(fmt.Sprintf(`"description": "%s"`, newDesc))
	res.AssertBodyContains(`"price": ` + newPrice)

	// Retreive a fresh copy of the item from the database and make sure
	// the updates are reflected in the database.
	updatedItem := &models.Item{}
	if err := zoom.ScanById(origItem.Id, updatedItem); err != nil {
		panic(err)
	}
	if updatedItem.Name != origItem.Name {
		t.Errorf("Updated name was incorrect. Expected %s but got %s", origItem.Name, updatedItem.Name)
	}
	if updatedItem.ImageUrl != origItem.ImageUrl {
		t.Errorf("Updated imageUrl was incorrect. Expected %s but got %s", origItem.ImageUrl, updatedItem.ImageUrl)
	}
	if updatedItem.Description != newDesc {
		t.Errorf("Updated description was incorrect. Expected %s but got %s", origItem.Description, newDesc)
	}
	newPriceFloat, err := strconv.ParseFloat(newPrice, 64)
	if err != nil {
		panic(err)
	}
	if updatedItem.Price != newPriceFloat {
		t.Errorf("Updated price was incorrect. Expected %s but got %s", origItem.Price, newPrice)
	}

	// CASE 1: Update the image file but not the name. This should result in the new file
	// replacing the old file, but the name staying the same.
	updateImage := updateItemRequest(rec, origItem.Id, map[string]string{}, redImage)
	res = rec.Do(updateImage)
	res.AssertOk()
	// Make sure the image url is unchanged
	res.AssertBodyContains(origItem.ImageUrl)

	// Make sure the image on s3 has the same name
	if !s3FileExists(origItem.ImageS3Path) {
		t.Errorf("The file does not exist or its name was changed on s3. Expected: %s", origItem.ImageS3Path)
	} else {
		// Make sure the contents of the new file are what we expect
		if calculateHashForS3File(origItem.ImageS3Path) != redImageHash {
			t.Error("The s3 file hash did not equal what we expected it to be (the red image hash). Therefore, image file on s3 was not updated.")
		}
	}

	// CASE 2: Update the image file *and* the name. We expect the old image file to be deleted
	// and a new one to be created using the new name.
	newName := "Test Item Update 1"
	updateNameAndImage := updateItemRequest(rec, origItem.Id, map[string]string{
		"name": newName,
	}, clearImage)
	res = rec.Do(updateNameAndImage)
	res.AssertOk()

	// Get the most up-to-date item from the database and check that our changes
	// were reflected in the database.
	updatedNameAndImageItem := &models.Item{}
	if err := zoom.ScanById(origItem.Id, updatedNameAndImageItem); err != nil {
		panic(err)
	}
	if updatedNameAndImageItem.Name != newName {
		t.Errorf("The item name was not updated. Expected %s but got %s", newName, updatedNameAndImageItem.Name)
	}
	if updatedNameAndImageItem.ImageS3Path == origItem.ImageS3Path {
		t.Error("The ImageS3Path property was not updated, but we expected it to be because the name changed.")
	}
	if updatedNameAndImageItem.ImageUrl == origItem.ImageUrl {
		t.Error("The ImageUrl property was not updated, but we expected it to be because the name changed.")
	}

	// Make sure the old file no longer exists
	if s3FileExists(origItem.ImageS3Path) {
		t.Errorf("The old file still exists at %s. Should be deleted since the name was changed.", origItem.ImageS3Path)
	}

	// Make sure the new file does exist
	if !s3FileExists(updatedNameAndImageItem.ImageS3Path) {
		t.Errorf("The new image file was not created at %s after the name was changed.", updatedNameAndImageItem.ImageS3Path)
	} else {
		// Make sure the contents of the new file are what we expect
		if calculateHashForS3File(updatedNameAndImageItem.ImageS3Path) != clearImageHash {
			t.Error("The s3 file hash did not equal what we expected it to be (the clear image hash). Therefore, image file on s3 was not updated.")
		}
	}

	// CASE 3: Update the item name but keep the image file the same. We expect the old image file
	// will be renamed (deleted and moved to the new location).
	newestName := "Test Item Update 2"
	updateNameOnly := updateItemRequest(rec, origItem.Id, map[string]string{
		"name": newestName,
	}, "")
	res = rec.Do(updateNameOnly)
	res.AssertOk()

	// Get the most up-to-date item from the database and check that our changes
	// were reflected in the database.
	updatedNameOnlyItem := &models.Item{}
	if err := zoom.ScanById(origItem.Id, updatedNameOnlyItem); err != nil {
		panic(err)
	}
	if updatedNameOnlyItem.Name != newestName {
		t.Errorf("The item name was not updated. Expected %s but got %s", newestName, updatedNameOnlyItem.Name)
	}
	if updatedNameOnlyItem.ImageS3Path == updatedNameAndImageItem.ImageS3Path {
		t.Error("The ImageS3Path property was not updated, but we expected it to be because the name changed.")
	}
	if updatedNameOnlyItem.ImageUrl == updatedNameAndImageItem.ImageUrl {
		t.Error("The ImageUrl property was not updated, but we expected it to be because the name changed.")
	}

	// Make sure the old file no longer exists
	if s3FileExists(updatedNameAndImageItem.ImageS3Path) {
		t.Errorf("The old file still exists at %s. Should be deleted since the name was changed.",
			updatedNameAndImageItem.ImageS3Path)
	}

	// Make sure the new file does exist
	if !s3FileExists(updatedNameOnlyItem.ImageS3Path) {
		t.Errorf("The new image file was not created at %s after the name was changed.", updatedNameOnlyItem.ImageS3Path)
	} else {
		// Make sure the contents of the new file are what we expect
		if calculateHashForS3File(updatedNameOnlyItem.ImageS3Path) != clearImageHash {
			t.Error("The s3 file hash did not equal what we expected it to be (the clear image hash). Therefore, image file on s3 was changed?")
		}
	}
}

func TestItemsDelete(t *testing.T) {
	t.Parallel()
	rec := fipple.NewRecorder(t, testUrl)

	// First create a test item
	deleteName := "Test Item Delete"
	deleteDesc := "An item for testing delete functionality."
	deletePrice := "0.99"
	createReq := createItemRequest(rec, map[string]string{
		"name":        deleteName,
		"description": deleteDesc,
		"price":       deletePrice,
	}, blueImage)
	rec.Do(createReq)

	// Get the item we created from the database
	item := &models.Item{}
	if err := zoom.NewQuery("Item").Filter("Name =", deleteName).ScanOne(item); err != nil {
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
	if count, err := zoom.NewQuery("Item").Filter("Name =", deleteName).Count(); err != nil {
		panic(err)
	} else if count != 0 {
		t.Error("Item was not deleted.")
	}

	// Make sure image was actually deleted from s3
	if s3FileExists(item.ImageS3Path) {
		t.Error("File was not deleted from s3.")
	}
}

func TestItemsIndex(t *testing.T) {
	rec := fipple.NewRecorder(t, testUrl)

	// First create two test items
	indexDescription := "This item should show up in the index."
	allItems := []*models.Item{
		createMockItem("Test Item Index 0", indexDescription, 99.01),
		createMockItem("Test Item Index 1", indexDescription, 99.02),
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

// createItemRequest creates and returns an http.Request with the given parameters, which
// will create an item when sent (e.g. with rec.Do).
func createItemRequest(rec *fipple.Recorder, fields map[string]string, imageFile string) *http.Request {
	// Set up the files that will be written to the form
	testImageFile, err := os.Open(imageFile)
	if err != nil {
		panic(err)
	}
	files := map[string]*os.File{
		"image": testImageFile,
	}

	// Create the request, add token, and return it
	req := rec.NewMultipartRequest("POST", "/items", fields, files)
	token, err := getAdminTestToken()
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	return req
}

// updateItemRequest creates and returns an http.Request with the given parameters, which
// will update an existing item when sent (e.g. with rec.Do). If you don't want to include the
// image file, just pass in a blank string.
func updateItemRequest(rec *fipple.Recorder, id string, fields map[string]string, imageFile string) *http.Request {
	// Set up the files that will be written to the form
	files := map[string]*os.File{}
	if imageFile != "" {
		testImageFile, err := os.Open(imageFile)
		if err != nil {
			panic(err)
		}
		files = map[string]*os.File{
			"image": testImageFile,
		}
	}

	// Create the request, add token, and return it
	path := fmt.Sprintf("/items/%s", id)
	req := rec.NewMultipartRequest("PUT", path, fields, files)
	token, err := getAdminTestToken()
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	return req
}
