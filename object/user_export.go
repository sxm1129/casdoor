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

package object

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
)

// ExportUsersToCSV exports all users for an organization to CSV format.
func ExportUsersToCSV(owner string) ([]byte, error) {
	users, err := GetUsers(owner)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %v", err)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Header row
	header := []string{
		"Name", "DisplayName", "Email", "Phone", "Type",
		"CreatedTime", "IsAdmin", "IsForbidden", "IsDeleted",
		"SignupApplication", "Score", "Ranking", "Tag",
		"Affiliation", "Title", "Region", "Language", "Gender",
		"Groups",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Data rows
	for _, user := range users {
		groups := strings.Join(user.Groups, ";")
		row := []string{
			user.Name,
			user.DisplayName,
			user.Email,
			user.Phone,
			user.Type,
			user.CreatedTime,
			boolToStr(user.IsAdmin),
			boolToStr(user.IsForbidden),
			boolToStr(user.IsDeleted),
			user.SignupApplication,
			fmt.Sprintf("%d", user.Score),
			fmt.Sprintf("%d", user.Ranking),
			user.Tag,
			user.Affiliation,
			user.Title,
			user.Region,
			user.Language,
			user.Gender,
			groups,
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
