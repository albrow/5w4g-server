package tests

import (
	"github.com/albrow/5w4g-server/lib"
	"github.com/albrow/fipple"
	"testing"
)

func TestAdminAuth(t *testing.T) {
	rec := fipple.NewRecorder(t, testUrl)

	// tests represents all the endpoints which should require admin authentication
	tests := []struct {
		method string
		path   string
	}{
		// Admin Users
		{"POST", "/admin_users"},
		{"GET", "/admin_users/foo"},
		{"GET", "/admin_users"},
		{"DELETE", "/admin_users/foo"},
		// Items
		{"POST", "/items"},
		{"PUT", "/items/foo"},
		{"DELETE", "/items/foo"},
		// Orders
		{"GET", "/orders"},
		{"GET", "/orders/foo"},
		{"PUT", "/orders/foo"},
		{"DELETE", "/orders/foo"},
	}

	for _, test := range tests {
		req := rec.NewRequest(test.method, test.path)
		res := rec.Do(req)
		res.AssertCode(401)
		res.AssertBodyContains(lib.ErrUnauthorized["error"])
	}
}
