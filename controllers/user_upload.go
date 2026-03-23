// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func saveFile(path string, file *multipart.File) (err error) {
	f, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, *file)
	if err != nil {
		return err
	}
	return nil
}

func (c *ApiController) UploadUsers() {
	if !c.IsAdmin() {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	userObj := c.getCurrentUser()
	if userObj == nil {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	userId := c.GetSessionUsername()
	owner, user, err := util.GetOwnerAndNameFromIdWithError(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	file, header, err := c.Ctx.Request.FormFile("file")
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	fileId := fmt.Sprintf("%s_%s_%s", owner, user, util.RemoveExt(header.Filename))
	path := util.GetUploadXlsxPath(fileId)
	defer os.Remove(path)
	err = saveFile(path, &file)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	affected, err := object.UploadUsers(owner, path, userObj, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if affected {
		c.ResponseOk()
	} else {
		c.ResponseError(c.T("general:Failed to import users"))
	}
}

// ExportUsers
// @Title ExportUsers
// @Tag User API
// @Description export users to CSV file
// @Param   owner    query    string  true        "organization name"
// @Success 200 {string} CSV file content
// @router /export-users [get]
func (c *ApiController) ExportUsers() {
	if !c.IsAdmin() {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	owner := c.Ctx.Input.Query("owner")
	if owner == "" {
		c.ResponseError(c.T("general:Missing parameter") + ": owner")
		return
	}

	csvData, err := object.ExportUsersToCSV(owner)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	filename := fmt.Sprintf("users_%s.csv", owner)
	c.Ctx.Output.Header("Content-Type", "text/csv; charset=utf-8")
	c.Ctx.Output.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	_, _ = c.Ctx.ResponseWriter.Write(csvData)
}
