package controllers

import (
	"github.com/unrolled/render"
	"net/http"
)

type AdminUserController struct{}

func (c *AdminUserController) SignIn(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	r.JSON(res, 200, map[string]interface{}{"token": "abcdefg"})
}
