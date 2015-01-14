package main

import (
	"fmt"
	"github.com/albrow/5w4g-server/config"
	"github.com/albrow/5w4g-server/controllers"
	"github.com/albrow/5w4g-server/models"
	"github.com/albrow/negroni-json-recovery"
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
		AllowHeaders:     []string{"Origin", "Content-Type", "X-Requested-With", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
	}))

	// Set up recovery middleware
	recovery.StackDepth = 3
	recovery.Formatter = func(errMsg string, stack []byte, file string, line int, fullMessages bool) interface{} {
		result := map[string]interface{}{
			"error": errMsg,
		}
		if fullMessages {
			result["lineNumber"] = fmt.Sprintf("%s:%d", file, line)
		}
		return result
	}
	n.Use(recovery.JSONRecovery(config.Env != "production"))

	// Define routes
	router := mux.NewRouter()

	// Admin Authentication
	adminTokens := controllers.AdminTokensController{}
	router.HandleFunc("/admin_users/sign_in", adminTokens.Create).Methods("POST")

	// Admin Users
	adminUsers := controllers.AdminUsersController{}
	router.HandleFunc("/admin_users", adminUsers.Create).Methods("POST")
	router.HandleFunc("/admin_users/{id}", adminUsers.Show).Methods("GET")
	router.HandleFunc("/admin_users", adminUsers.Index).Methods("GET")
	router.HandleFunc("/admin_users/{id}", adminUsers.Delete).Methods("DELETE")

	// Items
	items := controllers.ItemsController{}
	router.HandleFunc("/items", items.Create).Methods("POST")
	router.HandleFunc("/items", items.Index).Methods("GET")
	router.HandleFunc("/items/{id}", items.Show).Methods("GET")
	router.HandleFunc("/items/{id}", items.Update).Methods("PUT")
	router.HandleFunc("/items/{id}", items.Delete).Methods("DELETE")

	// Orders
	orders := controllers.OrdersController{}
	router.HandleFunc("/orders", orders.Create).Methods("POST")
	router.HandleFunc("/orders", orders.Index).Methods("GET")
	router.HandleFunc("/orders/{id}", orders.Show).Methods("GET")
	router.HandleFunc("/orders/{id}", orders.Update).Methods("PUT")
	router.HandleFunc("/orders/{id}", orders.Delete).Methods("DELETE")
	n.UseHandler(router)

	// Run
	n.Run(":" + config.Port)
}
