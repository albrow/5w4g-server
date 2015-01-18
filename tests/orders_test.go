package tests

import (
	"fmt"
	"github.com/albrow/5w4g-server/config"
	"github.com/albrow/5w4g-server/models"
	"github.com/albrow/fipple"
	"github.com/albrow/zoom"
	"strconv"
	"testing"
)

func TestOrdersCreate(t *testing.T) {
	rec := fipple.NewRecorder(t, testUrl)

	// First create some items which we can add to the order
	items := []*models.Item{}
	for i := 0; i < 3; i++ {
		iStr := strconv.Itoa(i)
		item := createMockItem("Order Test Item "+iStr, "This is item "+iStr, float64(i*10)+0.5)
		items = append(items, item)
	}

	// Construct a request to create an order
	orderItems := []map[string]interface{}{}
	for i, item := range items {
		orderItems = append(orderItems, map[string]interface{}{
			"itemId":   item.Id,
			"quantity": i + 1,
		})
	}
	orderEmail := "test@test.com"
	orderData := map[string]interface{}{
		"email": orderEmail,
		"items": orderItems,
	}
	req := rec.NewJSONRequest("POST", "/orders", orderData)

	// Send the request and check the response
	res := rec.Do(req)
	res.AssertOk()
	res.AssertBodyContains(fmt.Sprintf(`"email": "%s"`, orderEmail))
	for _, item := range items {
		res.AssertBodyContains(fmt.Sprintf(`"name": "%s"`, item.Name))
		res.AssertBodyContains(fmt.Sprintf(`"description": "%s"`, item.Description))
		res.AssertBodyContains(fmt.Sprintf(`"price": %0.1f`, item.Price))
	}

	// Check the database to make sure the order was actually created
	config.Init()
	models.Init()
	order := &models.Order{}
	if err := zoom.NewQuery("Order").Filter("Email =", orderEmail).ScanOne(order); err != nil {
		if _, ok := err.(*zoom.ModelNotFoundError); ok {
			t.Error("Order model was not saved in the database.")
		} else {
			panic(err)
		}
	}

	// Check that all the items associated with the order are correct
	if len(order.Items) != len(items) {
		t.Errorf("order.Items was not the right length. Expected %d but got %d", len(items), len(order.Items))
	}
	anyItemsMissing := false
	for _, expectedItem := range orderItems {
		itemFound := false
		expectedId, ok := expectedItem["itemId"].(string)
		if !ok {
			panic("Could not convert itemId to string!")
		}
		for _, gotOrderItem := range order.Items {
			if gotOrderItem.Item.Id == expectedId {
				itemFound = true
				expectedQuantity, ok := expectedItem["quantity"].(int)
				if !ok {
					panic("Could not convert quantity to int!")
				}
				if gotOrderItem.Quantity != expectedQuantity {
					t.Errorf("Incorrect quantity for item with id = %s. Expected %d but got %d",
						expectedId,
						expectedQuantity,
						gotOrderItem.Quantity)
				}
				break
			}
		}
		if !itemFound {
			anyItemsMissing = true
			t.Errorf("order.Items was missing expected item with id = %s", expectedId)
		}
	}
	if anyItemsMissing {
		gotIds := []string{}
		for _, orderItem := range order.Items {
			gotIds = append(gotIds, orderItem.Item.Id)
		}
		t.Errorf("order.Items consisted of the following item ids: %v", gotIds)
	}

	// TODO: test server-side validation errors
}
