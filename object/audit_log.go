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
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/casdoor/casdoor/util"
)

// AuditLog records field-level changes for compliance and debugging.
type AuditLog struct {
	Id          int    `xorm:"int notnull pk autoincr" json:"id"`
	CreatedTime string `xorm:"varchar(100) index" json:"createdTime"`

	// Who performed the action
	Actor     string `xorm:"varchar(200) index" json:"actor"`
	ActorIp   string `xorm:"varchar(100)" json:"actorIp"`

	// What was changed
	TargetType string `xorm:"varchar(100) index" json:"targetType"` // user, role, permission, organization
	TargetId   string `xorm:"varchar(200) index" json:"targetId"`
	Action     string `xorm:"varchar(50) index" json:"action"` // create, update, delete

	// Field-level diff (JSON)
	FieldChanges string `xorm:"mediumtext" json:"fieldChanges"`

	// Extra context
	LoginMethod string `xorm:"varchar(50)" json:"loginMethod"` // password, master_password, oauth, etc.
	RequestUri  string `xorm:"varchar(500)" json:"requestUri"`
}

// FieldChange represents a single field change in the audit log.
type FieldChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"oldValue"`
	NewValue interface{} `json:"newValue"`
}

func AddAuditLog(log *AuditLog) error {
	log.CreatedTime = util.GetCurrentTime()
	_, err := ormer.Engine.Insert(log)
	return err
}

func GetAuditLogs(targetType string, targetId string, limit int) ([]*AuditLog, error) {
	logs := []*AuditLog{}
	session := ormer.Engine.Desc("id")
	if targetType != "" {
		session = session.Where("target_type = ?", targetType)
	}
	if targetId != "" {
		session = session.Where("target_id = ?", targetId)
	}
	if limit > 0 {
		session = session.Limit(limit)
	}
	err := session.Find(&logs)
	return logs, err
}

func GetPaginationAuditLogs(offset, limit int, field, value, sortField, sortOrder string) ([]*AuditLog, error) {
	logs := []*AuditLog{}
	session := GetSession("", offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&logs)
	return logs, err
}

func GetAuditLogCount(field, value string) (int64, error) {
	session := GetSession("", -1, -1, field, value, "", "")
	return session.Count(&AuditLog{})
}

// GenerateFieldDiff compares two structs and returns a JSON string of changed fields.
// It uses reflection to compare exported fields, skipping fields that haven't changed.
func GenerateFieldDiff(oldObj, newObj interface{}) string {
	changes := []FieldChange{}

	oldVal := reflect.ValueOf(oldObj)
	newVal := reflect.ValueOf(newObj)

	if oldVal.Kind() == reflect.Ptr {
		oldVal = oldVal.Elem()
	}
	if newVal.Kind() == reflect.Ptr {
		newVal = newVal.Elem()
	}

	if oldVal.Type() != newVal.Type() {
		return "[]"
	}

	t := oldVal.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		// Skip large/internal fields
		name := field.Name
		if name == "Password" || name == "PasswordSalt" || name == "AccessSecret" || name == "TotpSecret" || name == "RecoveryCodes" {
			continue
		}

		oldF := oldVal.Field(i)
		newF := newVal.Field(i)

		if !reflect.DeepEqual(oldF.Interface(), newF.Interface()) {
			change := FieldChange{
				Field:    name,
				OldValue: fmt.Sprintf("%v", oldF.Interface()),
				NewValue: fmt.Sprintf("%v", newF.Interface()),
			}
			changes = append(changes, change)
		}
	}

	if len(changes) == 0 {
		return "[]"
	}

	data, err := json.Marshal(changes)
	if err != nil {
		return "[]"
	}
	return string(data)
}
