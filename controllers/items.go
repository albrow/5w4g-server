package controllers

import (
	"fmt"
	"github.com/albrow/5w4g-server/config"
	"github.com/albrow/5w4g-server/lib"
	"github.com/albrow/5w4g-server/models"
	"github.com/albrow/go-data-parser"
	"github.com/albrow/zoom"
	"github.com/gorilla/mux"
	"github.com/mitchellh/goamz/s3"
	"github.com/unrolled/render"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

type ItemsController struct{}

var acceptedImageExts = []string{"gif", "svg"}

func (c ItemsController) Create(res http.ResponseWriter, req *http.Request) {
	r := render.New()

	// Parse data from request
	itemData, err := data.Parse(req)
	if err != nil {
		panic(err)
	}

	// Validations
	val := itemData.Validator()
	val.Require("name")
	val.Require("price")
	val.Greater("price", 0.0)
	val.Require("description")
	val.RequireFile("image")
	val.AcceptFileExts("image", acceptedImageExts...)
	if itemData.Get("name") != "" {
		// Validate that name is unique
		count, err := zoom.NewQuery("Item").Filter("Name =", itemData.Get("name")).Count()
		if err != nil {
			panic(err)
		}
		if count != 0 {
			val.AddError("name", "that item name is already taken.")
		}
	}
	if val.HasErrors() {
		r.JSON(res, lib.StatusUnprocessableEntity, val.ErrorMap())
		return
	}

	// Create model with the attributes we have so far
	item := &models.Item{
		Name:        itemData.Get("name"),
		Price:       itemData.GetFloat("price"),
		Description: itemData.Get("description"),
	}

	// Upload the image to S3
	if imagePath, imageUrl, err := uploadImage(itemData.GetFile("image"), item.Name); err != nil {
		panic(err)
	} else {
		item.ImageS3Path = imagePath
		item.ImageUrl = imageUrl
	}

	// Save the item to the database
	if err := zoom.Save(item); err != nil {
		panic(err)
	}

	// Render response
	r.JSON(res, http.StatusOK, item)
}

func (c ItemsController) Show(res http.ResponseWriter, req *http.Request) {
	r := render.New()

	// Get the id from the url
	vars := mux.Vars(req)
	id := vars["id"]

	// Find item in the database
	item := &models.Item{}
	if err := zoom.ScanById(id, item); err != nil {
		panic(err)
	}

	// render response
	r.JSON(res, http.StatusOK, item)
}

func (c ItemsController) Update(res http.ResponseWriter, req *http.Request) {
	r := render.New()

	// Get the id from the url
	vars := mux.Vars(req)
	id := vars["id"]

	// Parse data from request
	itemData, err := data.Parse(req)
	if err != nil {
		panic(err)
	}

	// Validations
	val := itemData.Validator()
	if itemData.KeyExists("name") {
		val.Require("name").Message("name cannot be blank")
	}
	if itemData.KeyExists("price") {
		val.Require("price").Message("price cannot be blank")
		val.Greater("price", 0.0)
	}
	if itemData.KeyExists("description") {
		val.Require("description").Message("description cannot be blank")
	}
	if itemData.KeyExists("name") {
		// Validate that name is unique
		otherItem := &models.Item{}
		if err := zoom.NewQuery("Item").Filter("Name =", itemData.Get("name")).ScanOne(otherItem); err != nil {
			if _, ok := err.(*zoom.ModelNotFoundError); ok {
				// This means there was no model with the given name. That's fine.
			} else {
				// This means there was a problem connecting to the database
				panic(err)
			}
		} else {
			if otherItem.Id != id {
				// If the model with that name is the same model we're updating, that's
				// fine. We only want to return an error if that's not the case.
				val.AddError("name", "that item name is already taken.")
			}
		}
	}
	if itemData.KeyExists("image") {
		val.RequireFile("image") // Makes sure the file is not empty
		val.AcceptFileExts("image", acceptedImageExts...)
	}

	// Render validation errors if any
	if val.HasErrors() {
		r.JSON(res, lib.StatusUnprocessableEntity, val.ErrorMap())
		return
	}

	// Find item in the database
	item := &models.Item{}
	if err := zoom.ScanById(id, item); err != nil {
		panic(err)
	}

	// Update the item
	nameChanged := false
	if itemData.KeyExists("name") {
		nameChanged = item.Name != itemData.Get("name")
		item.Name = itemData.Get("name")
	}
	if itemData.KeyExists("description") {
		item.Description = itemData.Get("description")
	}
	if itemData.KeyExists("price") {
		item.Price = itemData.GetFloat("price")
	}

	// Handle different image upload cases
	switch {
	case itemData.FileExists("image") && !nameChanged:
		// New image provided, name of the image file should stay the same
		if imagePath, imageUrl, err := uploadImage(itemData.GetFile("image"), item.Name); err != nil {
			panic(err)
		} else {
			item.ImageS3Path = imagePath
			item.ImageUrl = imageUrl
		}
	case itemData.FileExists("image") && nameChanged:
		// We should delete the old image (which uses the old name)
		// and then upload the new one using the new item name
		if err := deleteImage(item.ImageS3Path); err != nil {
			panic(err)
		}
		if imagePath, imageUrl, err := uploadImage(itemData.GetFile("image"), item.Name); err != nil {
			panic(err)
		} else {
			item.ImageS3Path = imagePath
			item.ImageUrl = imageUrl
		}
	case !itemData.FileExists("image") && nameChanged:
		// We should rename the existing (old) image file since the
		// item name has been changed
		newPath, newUrl, err := renameImage(item.ImageS3Path, item.Name)
		if err != nil {
			panic(err)
		}
		item.ImageS3Path = newPath
		item.ImageUrl = newUrl
	}
	if err := zoom.Save(item); err != nil {
		panic(err)
	}

	// Render response
	r.JSON(res, http.StatusOK, item)
}

func (c ItemsController) Delete(res http.ResponseWriter, req *http.Request) {
	r := render.New()

	// Get the id from the url
	vars := mux.Vars(req)
	id := vars["id"]

	// Get the item from the database
	item := &models.Item{}
	if err := zoom.ScanById(id, item); err != nil {
		panic(err)
	}

	// Delete the image from S3
	deleteImage(item.ImageS3Path)

	// Delete from database
	if err := zoom.Delete(item); err != nil {
		panic(err)
	}

	// Render response
	r.JSON(res, http.StatusOK, struct{}{})
}

func (c ItemsController) Index(res http.ResponseWriter, req *http.Request) {
	r := render.New()

	// Find all admin users in the database
	var items []*models.Item
	if err := zoom.NewQuery("Item").Scan(&items); err != nil {
		panic(err)
	}

	// Render response
	r.JSON(res, http.StatusOK, items)
}

func calculateImageS3Path(itemName, filename string) string {
	imageFilename := url.QueryEscape(itemName)
	return fmt.Sprintf("items/%s%s", imageFilename, filepath.Ext(filename))
}

func calculateImageUrl(itemName, filename string) string {
	orig := calculateImageS3Path(itemName, filename)

	// In the url you use to actually get the image file, Amazon replaces "+" with
	// "%2B", so we'll do that too. WARNING: there may be other characters where this
	// happens too. The bug occurs because there are some characters that go's url.QueryEscape
	// that uses Amazon doesn't like even though they are technically safe for urls according to
	// spec
	return fmt.Sprintf("https://s3.amazonaws.com/%s/%s",
		config.Aws.BucketName,
		strings.Replace(orig, "+", "%2B", -1))
}

func uploadImage(fileHeader *multipart.FileHeader, itemName string) (imageOrigPath string, imageUrl string, e error) {
	// Get the image file from the header
	imageFile, err := fileHeader.Open()
	if err != nil {
		return "", "", err
	}

	// Get the raw bytes from the image file
	imageBytes, err := ioutil.ReadAll(imageFile)
	if err != nil {
		return "", "", err
	}
	// Get the mimetype of the image file
	imageType, _ := lib.GetImageMimeType(fileHeader.Filename)

	// Calculate and set original image path
	imageOrigPath = calculateImageS3Path(itemName, fileHeader.Filename)

	// Get the bucket instance
	bucket, err := lib.S3Bucket()
	if err != nil {
		return "", "", err
	}

	// Push the image file to the bucket
	if err := bucket.Put(imageOrigPath, imageBytes, imageType, s3.PublicRead); err != nil {
		return "", "", err
	}

	// Calculate and set image url
	imageUrl = calculateImageUrl(itemName, fileHeader.Filename)
	return imageOrigPath, imageUrl, nil
}

func deleteImage(path string) error {
	// Get the bucket instance
	bucket, err := lib.S3Bucket()
	if err != nil {
		return err
	}

	// Delete the file from the bucket
	if err := bucket.Del(path); err != nil {
		return err
	}
	return nil
}

func renameImage(oldPath string, newName string) (newPath string, newUrl string, e error) {
	// Get bucket
	bucket, err := lib.S3Bucket()
	if err != nil {
		return "", "", err
	}
	oldUrl, err := url.Parse(oldPath)
	oldFilename := filepath.Base(oldUrl.Path)
	newPath = calculateImageS3Path(newName, oldFilename)
	newUrl = calculateImageUrl(newName, oldFilename)

	// As far as I know the only way to do this with goamz is to
	// copy and then delete
	if err := bucket.Copy(oldPath, newPath, s3.PublicRead); err != nil {
		return "", "", err
	}
	if err := deleteImage(oldPath); err != nil {
		return "", "", err
	}
	return newPath, newUrl, nil
}
