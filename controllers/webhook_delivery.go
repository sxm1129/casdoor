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

import "github.com/casdoor/casdoor/object"

// GetWebhookDeliveries
// @Title GetWebhookDeliveries
// @Tag Webhook API
// @Description get webhook delivery records
// @Success 200 {array} object.WebhookDelivery
// @router /get-webhook-deliveries [get]
func (c *ApiController) GetWebhookDeliveries() {
	webhookId := c.Ctx.Input.Query("webhookId")
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")

	limitNum := parseIntStr(limit, 20)
	pageNum := parseIntStr(page, 1)
	offset := (pageNum - 1) * limitNum

	count, err := object.GetWebhookDeliveryCount(webhookId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	deliveries, err := object.GetWebhookDeliveries(webhookId, offset, limitNum)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(deliveries, count)
}
