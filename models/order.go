package models

import (
	"fmt"
)

type Order struct {
	Items []*OrderItem `json:"items"`
	Email string       `json:"email"`
	Identifier
}

// AddItem adds quantity of item to the order. It does not save the order, so
// you will need to do so if you want the changes to persist in the database.
// AddItem will return an error if the item you are attempting to add does not
// have an Id. If the item you are attempting to add already exists in o, AddItem
// will add quantity to the existing OrderItem.Quantity.
func (o *Order) AddItem(item *Item, quantity int) error {
	if item.Id == "" {
		return fmt.Errorf("Cannot add item with an empty Id: %+v", *item)
	}

	// Check if the item already exists for this order
	var existingOrderItem *OrderItem
	for _, existing := range o.Items {
		if existing.Item.Id == item.Id {
			existingOrderItem = existing
			break
		}
	}
	if existingOrderItem == nil {
		// If there *is not* an existing item, create one and add it to the order
		orderItem := &OrderItem{
			Item:     item,
			Quantity: quantity,
		}
		o.Items = append(o.Items, orderItem)
	} else {
		// If there *is* an existing item, add quantity to OrderItem.Quantity
		existingOrderItem.Quantity += quantity
	}
	return nil
}
