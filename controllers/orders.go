package controllers

import (
	"net/http"
)

type OrdersController struct{}

func (o OrdersController) Create(res http.ResponseWriter, req *http.Request) {
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
