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

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetPortalProfile
// @Title GetPortalProfile
// @Tag Portal API
// @Description get current user's profile for the self-service portal
// @Success 200 {object} object.User
// @router /get-portal-profile [get]
func (c *ApiController) GetPortalProfile() {
	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError("Please sign in first")
		return
	}

	owner, name := util.GetOwnerAndNameFromIdNoCheck(userId)
	user, err := object.GetUser(owner + "/" + name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError("User not found")
		return
	}

	user.Password = ""
	user.TotpSecret = ""
	user.RecoveryCodes = nil

	c.ResponseOk(user)
}

// UpdatePortalProfile
// @Title UpdatePortalProfile
// @Tag Portal API
// @Description update current user's own profile (restricted fields)
// @Param body body object.User true "User object"
// @Success 200 {object} controllers.Response
// @router /update-portal-profile [post]
func (c *ApiController) UpdatePortalProfile() {
	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError("Please sign in first")
		return
	}

	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	owner, name := util.GetOwnerAndNameFromIdNoCheck(userId)
	if user.Owner != owner || user.Name != name {
		c.ResponseError("You can only update your own profile")
		return
	}

	columns := []string{
		"display_name", "avatar", "email", "phone", "country_code",
		"region", "location", "address", "affiliation", "title",
		"homepage", "bio", "language",
	}

	success, err := object.UpdateUser(user.GetId(), &user, columns, false)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(success)
}

// GetPortalSessions
// @Title GetPortalSessions
// @Tag Portal API
// @Description get current user's active sessions
// @Success 200 {array} object.Session
// @router /get-portal-sessions [get]
func (c *ApiController) GetPortalSessions() {
	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError("Please sign in first")
		return
	}

	owner, _ := util.GetOwnerAndNameFromIdNoCheck(userId)
	sessions, err := object.GetSessions(owner)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(sessions)
}
