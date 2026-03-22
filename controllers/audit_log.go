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

// GetAuditLogs
// @Title GetAuditLogs
// @Tag Audit API
// @Description get audit logs
// @Success 200 {array} object.AuditLog
// @router /get-audit-logs [get]
func (c *ApiController) GetAuditLogs() {
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		targetType := c.Ctx.Input.Query("targetType")
		targetId := c.Ctx.Input.Query("targetId")
		logs, err := object.GetAuditLogs(targetType, targetId, 100)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk(logs)
		return
	}

	limitNum := parseIntStr(limit, 10)
	pageNum := parseIntStr(page, 1)
	offset := (pageNum - 1) * limitNum

	count, err := object.GetAuditLogCount(field, value)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	logs, err := object.GetPaginationAuditLogs(offset, limitNum, field, value, sortField, sortOrder)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(logs, count)
}

func parseIntStr(s string, defaultVal int) int {
	var n int
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			n = n*10 + int(ch-'0')
		}
	}
	if n == 0 {
		return defaultVal
	}
	return n
}
