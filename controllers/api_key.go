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
)

// GetApiKeys
// @Title GetApiKeys
// @Tag ApiKey API
// @Description get API keys for an owner
// @Success 200 {array} object.ApiKey
// @router /get-api-keys [get]
func (c *ApiController) GetApiKeys() {
	owner := c.Ctx.Input.Query("owner")
	if owner == "" {
		owner = c.GetSessionUsername()
	}

	keys, err := object.GetApiKeys(owner)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(keys)
}

// AddApiKey
// @Title AddApiKey
// @Tag ApiKey API
// @Description add a new API key
// @Param body body object.ApiKey true "ApiKey object"
// @Success 200 {object} controllers.Response
// @router /add-api-key [post]
func (c *ApiController) AddApiKey() {
	var apiKey object.ApiKey
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &apiKey)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	success, rawKey, err := object.AddApiKey(&apiKey)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if success {
		c.ResponseOk(rawKey)
	} else {
		c.ResponseError("Failed to create API key")
	}
}

// DeleteApiKey
// @Title DeleteApiKey
// @Tag ApiKey API
// @Description delete an API key
// @Param body body object.ApiKey true "ApiKey object"
// @Success 200 {object} controllers.Response
// @router /delete-api-key [post]
func (c *ApiController) DeleteApiKey() {
	var apiKey object.ApiKey
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &apiKey)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	success, err := object.DeleteApiKey(apiKey.Owner, apiKey.Name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(success)
}
