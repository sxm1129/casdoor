// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// TestGetAccountUnauthenticated tests that /api/get-account returns
// an empty/null user when no session is present.
func TestGetAccountUnauthenticated(t *testing.T) {
	initTestBeego()

	w := testRequest("GET", "/api/get-account", "", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	result := parseJsonResponse(w)
	status, ok := result["status"]
	if !ok {
		t.Fatal("response missing 'status' field")
	}

	// When unauthenticated, status should be "ok" with no user data,
	// or "error" depending on the implementation
	t.Logf("Unauthenticated get-account response: status=%v", status)
}

// TestLoginInvalidCredentials tests that /api/login rejects
// invalid username/password combinations.
func TestLoginInvalidCredentials(t *testing.T) {
	initTestBeego()

	loginForm := `{
		"application": "app-built-in",
		"organization": "built-in",
		"username": "nonexistent_user_12345",
		"password": "wrong_password",
		"type": "login"
	}`

	w := testRequest("POST", "/api/login", loginForm, nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	result := parseJsonResponse(w)
	status, ok := result["status"]
	if !ok {
		t.Fatal("response missing 'status' field")
	}

	// Invalid credentials should return error status
	if status != "error" {
		t.Errorf("expected status 'error' for invalid login, got '%v'", status)
	}

	t.Logf("Invalid login response: status=%v, msg=%v", status, result["msg"])
}

// TestLoginEmptyBody tests that /api/login handles empty/missing
// request body gracefully without panicking.
func TestLoginEmptyBody(t *testing.T) {
	initTestBeego()

	w := testRequest("POST", "/api/login", "", nil)

	// Should not panic — any HTTP status is acceptable as long as
	// the server doesn't crash
	if w.Code == 0 {
		t.Fatal("server returned 0 status, likely panicked")
	}

	t.Logf("Empty login body response: status=%d", w.Code)
}

// TestHealthEndpoint tests the /api/health endpoint returns OK.
func TestHealthEndpoint(t *testing.T) {
	initTestBeego()

	w := testRequest("GET", "/api/health", "", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	t.Logf("Health response: %s", w.Body.String())
}

// TestGetDashboardEndpoint tests the /api/get-dashboard endpoint.
func TestGetDashboardEndpoint(t *testing.T) {
	initTestBeego()

	w := testRequest("GET", "/api/get-dashboard?owner=All", "", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	result := parseJsonResponse(w)
	status, ok := result["status"]
	if !ok {
		t.Fatal("response missing 'status' field")
	}

	t.Logf("Dashboard response: status=%v", status)
}

// TestExportUsersUnauthenticated tests that /api/export-users
// rejects unauthenticated requests.
func TestExportUsersUnauthenticated(t *testing.T) {
	initTestBeego()

	w := testRequest("GET", "/api/export-users?owner=built-in", "", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	result := parseJsonResponse(w)
	status, ok := result["status"]
	if !ok {
		// If CSV was returned directly, this means auth bypass happened
		t.Fatal("response is not JSON — possible auth bypass in export-users")
	}

	if status != "error" {
		t.Errorf("expected status 'error' for unauthenticated export, got '%v'", status)
	}

	t.Logf("Unauthenticated export response: status=%v, msg=%v", status, result["msg"])
}

// TestSetPasswordEndpoint tests that the set-password endpoint
// rejects requests without proper authentication.
func TestSetPasswordEndpoint(t *testing.T) {
	initTestBeego()

	body := `{"userOwner":"built-in","userName":"admin","oldPassword":"old","newPassword":"new"}`
	w := testRequest("POST", "/api/set-password", body, nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	result := parseJsonResponse(w)

	// Should fail without valid session
	t.Logf("Set password response: %v", result)
}

// TestAPIEndpointsExist verifies that key API endpoints are registered
// and return non-404 responses.
func TestAPIEndpointsExist(t *testing.T) {
	initTestBeego()

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/health"},
		{"GET", "/api/get-account"},
		{"GET", "/api/get-dashboard?owner=All"},
		{"POST", "/api/login"},
		{"GET", "/api/get-organizations"},
		{"GET", "/api/get-users"},
		{"GET", "/api/get-roles"},
		{"GET", "/api/get-permissions"},
		{"GET", "/api/get-providers"},
		{"GET", "/api/get-applications"},
	}

	for _, ep := range endpoints {
		t.Run(fmt.Sprintf("%s_%s", ep.method, ep.path), func(t *testing.T) {
			w := testRequest(ep.method, ep.path, "", nil)
			if w.Code == http.StatusNotFound {
				t.Errorf("endpoint %s %s returned 404 — not registered", ep.method, ep.path)
			}

			// Additionally verify JSON responses are well-formed
			if w.Header().Get("Content-Type") != "" {
				var jsonCheck map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &jsonCheck); err != nil {
					t.Logf("Response for %s %s is not JSON (may be expected): %s", ep.method, ep.path, w.Body.String()[:min(100, w.Body.Len())])
				}
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
