package main

import (
	"github.com/albrow/5w4g/config"
	"github.com/albrow/5w4g/controllers"
	"github.com/albrow/5w4g/models"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/martini-contrib/cors"
)

func main() {
	// Initialize app
	config.Init()
	models.Init()

	// Add middleware
	n := negroni.New(negroni.NewLogger())
	n.UseHandler(cors.Allow(&cors.Options{
		AllowOrigins:     config.AllowOrigins,
		AllowMethods:     []string{"GET", "POST", "DELETE", "PUT", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
	}))

	// Define routes
	router := mux.NewRouter()
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminUsers := controllers.AdminUserController{}
	adminRouter.HandleFunc("/sign_in", adminUsers.SignIn).Methods("POST")
	n.UseHandler(router)

	// Run
	n.Run(":" + config.Port)
}
