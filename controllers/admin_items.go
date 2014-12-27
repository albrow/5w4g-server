package controllers

import (
	"github.com/albrow/5w4g-server/lib"
	"github.com/albrow/5w4g-server/models"
	"github.com/albrow/go-data-parser"
	"github.com/albrow/zoom"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"net/http"
)

type AdminItemsController struct{}

func (c AdminItemsController) Create(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	// Make sure we're signed in
	if currentUser := lib.CurrentAdminUser(req); currentUser == nil {
		r.JSON(res, 401, lib.ErrUnauthorized)
		return
	}

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

	// We'll check for the image file separately since go-data-parser doesn't support
	// this yet.
	_, _, err = req.FormFile("image")
	if err != nil {
		if err == http.ErrMissingFile {
			val.AddError("image", "image is required.")
		}
	}

	if val.HasErrors() {
		errors := map[string]interface{}{
			"errors": val.ErrorMap(),
		}
		r.JSON(res, 422, errors)
		return
	}

	// Create model and save to database
	item := &models.Item{
		Name: itemData.Get("name"),
		// TODO: upload the actual file to S3 and set the actual url for the image file
		ImageUrl:    "http://lorempixel.com/800/600/cats/",
		Price:       itemData.GetFloat("price"),
		Description: itemData.Get("description"),
	}
	if err := zoom.Save(item); err != nil {
		panic(err)
	}

	// Render response
	jsonData := map[string]interface{}{
		"item": item,
	}
	r.JSON(res, 200, jsonData)
}

func (c AdminItemsController) Show(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	// Make sure we're signed in
	if currentUser := lib.CurrentAdminUser(req); currentUser == nil {
		r.JSON(res, 401, lib.ErrUnauthorized)
		return
	}

	// Get the id from the url
	vars := mux.Vars(req)
	id, found := vars["id"]
	if !found {
		r.JSON(res, 422, lib.NewJsonError("Missing required url parameter: id"))
		return
	}

	// Find item in the database
	item := &models.Item{}
	if err := zoom.ScanById(id, item); err != nil {
		panic(err)
	}

	// render response
	jsonData := map[string]interface{}{"item": item}
	r.JSON(res, 200, jsonData)
}

func (c AdminItemsController) Update(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	// Make sure we're signed in
	if currentUser := lib.CurrentAdminUser(req); currentUser == nil {
		r.JSON(res, 401, lib.ErrUnauthorized)
		return
	}

	// Get the id from the url
	vars := mux.Vars(req)
	id, found := vars["id"]
	if !found {
		r.JSON(res, 422, lib.NewJsonError("Missing required url parameter: id"))
		return
	}

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
	// TODO: check for a valid image file
	if val.HasErrors() {
		errors := map[string]interface{}{
			"errors": val.ErrorMap(),
		}
		r.JSON(res, 422, errors)
		return
	}

	// Find item in the database
	item := &models.Item{}
	if err := zoom.ScanById(id, item); err != nil {
		panic(err)
	}

	// Update the item
	// TODO: update the image for the item if a new image file was provided
	if itemData.KeyExists("name") {
		item.Name = itemData.Get("name")
	}
	if itemData.KeyExists("description") {
		item.Description = itemData.Get("description")
	}
	if itemData.KeyExists("price") {
		item.Price = itemData.GetFloat("price")
	}
	if err := zoom.Save(item); err != nil {
		panic(err)
	}

	// Render response
	jsonData := map[string]interface{}{"item": item}
	r.JSON(res, 200, jsonData)
}

func (c AdminItemsController) Delete(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	// Make sure we're signed in
	if currentUser := lib.CurrentAdminUser(req); currentUser == nil {
		r.JSON(res, 401, lib.ErrUnauthorized)
		return
	}

	// Get the id from the url
	vars := mux.Vars(req)
	id, found := vars["id"]
	if !found {
		r.JSON(res, 422, lib.NewJsonError("Missing required url parameter: id"))
		return
	}

	// Delete from database
	if err := zoom.DeleteById("Item", id); err != nil {
		panic(err)
	}

	// Render response
	r.JSON(res, 200, struct{}{})
}

func (c AdminItemsController) Index(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	// Make sure we're signed in
	if currentUser := lib.CurrentAdminUser(req); currentUser == nil {
		r.JSON(res, 401, lib.ErrUnauthorized)
		return
	}

	// Find all admin users in the database
	var items []*models.Item
	if err := zoom.NewQuery("Item").Scan(&items); err != nil {
		panic(err)
	}

	// Render response
	jsonData := map[string]interface{}{"items": items}
	r.JSON(res, 200, jsonData)
}
