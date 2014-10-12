package main

import (
	"github.com/albrow/5w4g-server/config"
	"github.com/albrow/5w4g-server/controllers"
	"github.com/albrow/5w4g-server/models"
	"github.com/codegangsta/negroni"
	"github.com/goincremental/negroni-sessions"
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
	store := sessions.NewCookieStore(config.Secret)
	n.Use(sessions.Sessions("5w4g_session", store))

	// Define routes
	router := mux.NewRouter()
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminUsers := controllers.AdminUserController{}
	adminRouter.HandleFunc("/sign_in", adminUsers.SignIn).Methods("POST")
	n.UseHandler(router)

	// Run
	n.Run(":" + config.Port)
}
