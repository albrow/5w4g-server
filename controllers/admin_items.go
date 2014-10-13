package controllers

import (
	"github.com/albrow/5w4g-server/models"
	"github.com/albrow/go-data-parser"
	"github.com/albrow/zoom"
	"github.com/unrolled/render"
	"net/http"
)

type AdminItemsController struct{}

func (c AdminItemsController) Create(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	// Make sure we're signed in
	if currentUser := CurrentAdminUser(req); currentUser == nil {
		jsonData := map[string]interface{}{
			"errors": map[string][]string{
				"error": []string{"You need to be signed in to do that!"},
			},
		}
		r.JSON(res, 401, jsonData)
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
	val.Require("imageUrl")
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
	if val.HasErrors() {
		errors := map[string]interface{}{
			"errors": val.ErrorMap(),
		}
		r.JSON(res, 422, errors)
		return
	}

	// Create model and save to database
	item := &models.Item{
		Name:        itemData.Get("name"),
		ImageUrl:    itemData.Get("imageUrl"),
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

func (c AdminItemsController) Index(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	// Make sure we're signed in
	if currentUser := CurrentAdminUser(req); currentUser == nil {
		jsonData := map[string]interface{}{
			"errors": map[string][]string{
				"error": []string{"You need to be signed in to do that!"},
			},
		}
		r.JSON(res, 401, jsonData)
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
