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
		AllowHeaders:     []string{"Origin", "Content-Type", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
	}))
	store := sessions.NewCookieStore(config.Secret)
	n.Use(sessions.Sessions("5w4g_session", store))

	// Set up recovery middleware
	recovery.StackDepth = 3
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
	adminTokens := controllers.AdminTokensController{}
	adminRouter.HandleFunc("/sign_in", adminTokens.Create).Methods("POST")
	adminUsers := controllers.AdminUsersController{}
	adminRouter.HandleFunc("/users", adminUsers.Create).Methods("POST")
	adminRouter.HandleFunc("/users/{id}", adminUsers.Show).Methods("GET")
	adminRouter.HandleFunc("/users", adminUsers.Index).Methods("GET")
	adminRouter.HandleFunc("/users/{id}", adminUsers.Delete).Methods("DELETE")
	adminItems := controllers.AdminItemsController{}
	adminRouter.HandleFunc("/items", adminItems.Create).Methods("POST")
	adminRouter.HandleFunc("/items", adminItems.Index).Methods("GET")
	adminRouter.HandleFunc("/items/{id}", adminItems.Show).Methods("GET")
	adminRouter.HandleFunc("/items/{id}", adminItems.Update).Methods("PUT")
	adminRouter.HandleFunc("/items/{id}", adminItems.Delete).Methods("DELETE")
	n.UseHandler(router)

	// Run
	n.Run(":" + config.Port)
}
