package main

import (
	"fmt"
	"github.com/albrow/5w4g-server/config"
	"github.com/albrow/5w4g-server/controllers"
	"github.com/albrow/5w4g-server/lib"
	"github.com/albrow/5w4g-server/models"
	"github.com/albrow/negroni-json-recovery"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/martini-contrib/cors"
	"github.com/unrolled/render"
	"net/http"
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
	router.HandleFunc("/admin_users", RequireAdmin(adminUsers.Create)).Methods("POST")
	router.HandleFunc("/admin_users/{id}", RequireAdmin(adminUsers.Show)).Methods("GET")
	router.HandleFunc("/admin_users", RequireAdmin(adminUsers.Index)).Methods("GET")
	router.HandleFunc("/admin_users/{id}", RequireAdmin(adminUsers.Delete)).Methods("DELETE")

	// Items
	items := controllers.ItemsController{}
	router.HandleFunc("/items", RequireAdmin(items.Create)).Methods("POST")
	router.HandleFunc("/items", items.Index).Methods("GET")
	router.HandleFunc("/items/{id}", items.Show).Methods("GET")
	router.HandleFunc("/items/{id}", RequireAdmin(items.Update)).Methods("PUT")
	router.HandleFunc("/items/{id}", RequireAdmin(items.Delete)).Methods("DELETE")

	// Orders
	orders := controllers.OrdersController{}
	router.HandleFunc("/orders", orders.Create).Methods("POST")
	router.HandleFunc("/orders", RequireAdmin(orders.Index)).Methods("GET")
	router.HandleFunc("/orders/{id}", RequireAdmin(orders.Show)).Methods("GET")
	router.HandleFunc("/orders/{id}", RequireAdmin(orders.Update)).Methods("PUT")
	router.HandleFunc("/orders/{id}", RequireAdmin(orders.Delete)).Methods("DELETE")

	// Start the server
	n.UseHandler(router)
	n.Run(":" + config.Port)
}

// RequireAdmin is a middleware-like function that wraps around an http.HandlerFunc.
// It checks for the presence of a valid JWT in the header of the request. If the token
// is valid, it calls next. If the token wasn't provided or is invalid, it writes a
// 401 error to res and returns without calling next.
func RequireAdmin(next http.HandlerFunc) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		// If an admin user is not signed in, print an error and don't continue
		if currentUser := lib.CurrentAdminUser(req); currentUser == nil {
			r := render.New()
			r.JSON(res, 401, lib.ErrUnauthorized)
			return
		}
		// Otherwise, continue down the middleware chain by calling next
		next(res, req)
	}
}
