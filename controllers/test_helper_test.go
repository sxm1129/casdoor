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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/beego/beego/v2/server/web"
	"github.com/casdoor/casdoor/object"
)

var initOnce sync.Once

// initTestBeego initializes beego framework, DB, and routes for controller tests.
// Uses sync.Once to ensure single initialization across all tests.
// Routes are registered inline to avoid import cycle with routers package.
func initTestBeego() {
	initOnce.Do(func() {
		object.InitConfig()

		// Register a subset of routes needed for testing
		// (cannot import routers package due to circular dependency)
		web.Router("/api/login", &ApiController{}, "POST:Login")
		web.Router("/api/get-account", &ApiController{}, "GET:GetAccount")
		web.Router("/api/logout", &ApiController{}, "GET,POST:Logout")
		web.Router("/api/health", &ApiController{}, "GET:Health")
		web.Router("/api/get-dashboard", &ApiController{}, "GET:GetDashboard")
		web.Router("/api/export-users", &ApiController{}, "GET:ExportUsers")
		web.Router("/api/set-password", &ApiController{}, "POST:SetPassword")
		web.Router("/api/get-organizations", &ApiController{}, "GET:GetOrganizations")
		web.Router("/api/get-users", &ApiController{}, "GET:GetUsers")
		web.Router("/api/get-roles", &ApiController{}, "GET:GetRoles")
		web.Router("/api/get-permissions", &ApiController{}, "GET:GetPermissions")
		web.Router("/api/get-providers", &ApiController{}, "GET:GetProviders")
		web.Router("/api/get-applications", &ApiController{}, "GET:GetApplications")
	})
}

// testRequest creates an HTTP request and returns the recorded response.
func testRequest(method, path string, body string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req, _ := http.NewRequest(method, path, bodyReader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	web.BeeApp.Handlers.ServeHTTP(w, req)
	return w
}

// parseJsonResponse parses a JSON response body into a map.
func parseJsonResponse(w *httptest.ResponseRecorder) map[string]interface{} {
	var result map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &result)
	return result
}
