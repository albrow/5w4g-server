package main

import (
	"fmt"
	"github.com/albrow/5w4g/config"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	config.Init()

	n := negroni.New(negroni.NewLogger())
	router := mux.NewRouter()
	router.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprint(res, "Hello!")
	})
	n.UseHandler(router)

	n.Run(":" + config.Port)
}
