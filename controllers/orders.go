package controllers

import (
	"fmt"
	"github.com/albrow/5w4g-server/models"
	"github.com/albrow/go-data-parser"
	"github.com/albrow/zoom"
	"github.com/unrolled/render"
	"net/http"
)

type OrdersController struct{}

type orderItemDatum struct {
	ItemId   string `json:"itemId"`
	Quantity int    `json:"quantity"`
}

func (o OrdersController) Create(res http.ResponseWriter, req *http.Request) {
	r := render.New()

	// Make sure we're getting JSON
	if req.Header.Get("Content-Type") != "application/json" {
		msg := fmt.Sprintf(`Unsupported Content-Type: "%s". Expected "application/json".`, req.Header.Get("Content-Type"))
		r.JSON(res, http.StatusUnsupportedMediaType, map[string]string{
			"error": msg,
		})
	}

	// Parse data from the request.
	orderData, err := data.Parse(req)
	if err != nil {
		panic(err)
	}

	// Validations
	val := orderData.Validator()
	val.Require("email")
	val.Require("items")
	if val.HasErrors() {
		r.JSON(res, 422, val.ErrorMap())
		return
	}

	// Get and unmarshall the items key
	oiData := []orderItemDatum{}
	if err := orderData.GetAndUnmarshalJSON("items", &oiData); err != nil {
		panic(err)
	}

	// Get all the items by their id from the database.
	itemIds := make([]string, len(oiData))
	for i, datum := range oiData {
		if datum.ItemId == "" {
			// Return a validation error if any itemId parameters are blank
			msg := fmt.Sprintf("items[%d] had a blank itemId. itemId is required for each item.", i)
			r.JSON(res, 422, map[string]string{"items": msg})
			return
		}
		if (datum.Quantity <= 0) || (datum.Quantity >= 1e4) {
			// Return a validation error if any quantity parameters are
			// out of range.
			msg := fmt.Sprintf("items[%d] had an invalid quantity. quantity must be between 0 and 10,000.", i)
			r.JSON(res, 422, map[string]string{"items": msg})
			return
		}
		itemIds[i] = datum.ItemId
	}
	// Use MScanById to get all the items in one go
	items := make([]*models.Item, len(itemIds))
	if err := zoom.MScanById(itemIds, &items); err != nil {
		if _, ok := err.(*zoom.KeyNotFoundError); ok {
			// This means the itemId was invalid. Return a validation error.
			msg := fmt.Sprintf("One of the items had an invalid itemId. %s.", err.Error())
			r.JSON(res, 422, map[string]string{"items": msg})
			return
		} else {
			// For any other error, panic
			panic(err)
		}
	}

	// Create the Order model and add each item to the order
	order := &models.Order{
		Email: orderData.Get("email"),
	}
	for i, datum := range oiData {
		order.AddItem(items[i], datum.Quantity)
	}
	// Save all the OrderItems in one go using MSave
	if err := zoom.MSave(zoom.Models(order.Items)); err != nil {
		panic(err)
	}
	// Then save the order itself
	if err := zoom.Save(order); err != nil {
		panic(err)
	}

	// Return the Order
	r.JSON(res, 200, order)
}

func (o OrdersController) Show(res http.ResponseWriter, req *http.Request) {
	panic("Orders.Show not yet implemented!")
}

func (o OrdersController) Update(res http.ResponseWriter, req *http.Request) {
	panic("Orders.Update not yet implemented!")
}

func (o OrdersController) Delete(res http.ResponseWriter, req *http.Request) {
	panic("Orders.Delete not yet implemented!")
}

func (o OrdersController) Index(res http.ResponseWriter, req *http.Request) {
	panic("Orders.Index not yet implemented!")
}
