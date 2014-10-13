package main

import (
	"fmt"
	"github.com/albrow/5w4g-server/config"
	"github.com/albrow/5w4g-server/controllers"
	"github.com/albrow/5w4g-server/models"
	"github.com/albrow/negroni-json-recovery"
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
	recovery.Formatter = func(errMsg string, stack []byte, file string, line int, fullMessages bool) interface{} {
		result := map[string]interface{}{
			"errors": map[string][]string{
				"error": []string{errMsg},
			},
		}
		if fullMessages {
			result["lineNumber"] = fmt.Sprintf("%s:%d", file, line)
		}
		return result
	}
	n.Use(recovery.JSONRecovery(config.Env != "production"))

	// Define routes
	router := mux.NewRouter()
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminSessions := controllers.AdminSessionsController{}
	adminRouter.HandleFunc("/sessions", adminSessions.Create).Methods("POST")
	adminRouter.HandleFunc("/sessions", adminSessions.Delete).Methods("DELETE")
	adminRouter.HandleFunc("/sessions", adminSessions.Show).Methods("GET")
	adminUsers := controllers.AdminUsersController{}
	adminRouter.HandleFunc("/users", adminUsers.Create).Methods("POST")
	adminRouter.HandleFunc("/users", adminUsers.Index).Methods("GET")
	adminRouter.HandleFunc("/users/{id}", adminUsers.Delete).Methods("DELETE")
	adminItems := controllers.AdminItemsController{}
	adminRouter.HandleFunc("/items", adminItems.Create).Methods("POST")
	adminRouter.HandleFunc("/items", adminItems.Index).Methods("GET")
	n.UseHandler(router)

	// Run
	n.Run(":" + config.Port)
}
