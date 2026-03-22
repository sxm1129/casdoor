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
	"time"

	"github.com/casdoor/casdoor/cred"
	"github.com/casdoor/casdoor/util"
)

// PasswordHistory tracks previous password hashes for a user to prevent reuse.
type PasswordHistory struct {
	Id           int64  `xorm:"pk autoincr" json:"id"`
	Owner        string `xorm:"varchar(100) index" json:"owner"`
	Name         string `xorm:"varchar(255) index" json:"name"`
	PasswordHash string `xorm:"varchar(200)" json:"-"`
	PasswordSalt string `xorm:"varchar(100)" json:"-"`
	PasswordType string `xorm:"varchar(100)" json:"-"`
	CreatedTime  string `xorm:"varchar(100)" json:"createdTime"`
}

const defaultPasswordHistoryLimit = 5

// AddPasswordHistory records a password hash for history checking.
func AddPasswordHistory(user *User) error {
	record := &PasswordHistory{
		Owner:        user.Owner,
		Name:         user.Name,
		PasswordHash: user.Password,
		PasswordSalt: user.PasswordSalt,
		PasswordType: user.PasswordType,
		CreatedTime:  util.GetCurrentTime(),
	}

	_, err := ormer.Engine.Insert(record)
	return err
}

// CheckPasswordAgainstHistory checks if a new password matches any of the last N
// password history entries. Returns true if the password is NOT in history (safe to use).
func CheckPasswordAgainstHistory(user *User, newPassword string, organization *Organization, historyLimit int) (bool, error) {
	if historyLimit <= 0 {
		historyLimit = defaultPasswordHistoryLimit
	}

	histories := []*PasswordHistory{}
	err := ormer.Engine.Where("owner = ? AND name = ?", user.Owner, user.Name).
		Desc("id").
		Limit(historyLimit).
		Find(&histories)
	if err != nil {
		return true, err // On error, allow the password change (fail open)
	}

	for _, h := range histories {
		credManager := cred.GetCredManager(h.PasswordType)
		if credManager == nil {
			continue
		}
		// Check against both the history-stored salt and the organization salt
		if credManager.IsPasswordCorrect(newPassword, h.PasswordHash, h.PasswordSalt) ||
			credManager.IsPasswordCorrect(newPassword, h.PasswordHash, organization.PasswordSalt) {
			return false, nil // Password was used before
		}
	}

	return true, nil
}

// IsPasswordExpired checks if a user's password has exceeded the organization's expiration policy.
// Returns true if the password IS expired and needs to be changed.
func IsPasswordExpired(user *User, organization *Organization) bool {
	if organization == nil || organization.PasswordExpireDays <= 0 {
		return false
	}

	if user.LastChangePasswordTime == "" {
		// If no password change time recorded, consider it expired
		// (forces user to set a new password on next login)
		return true
	}

	lastChange, err := time.Parse("2006-01-02T15:04:05+08:00", user.LastChangePasswordTime)
	if err != nil {
		// Try RFC3339 format
		lastChange, err = time.Parse(time.RFC3339, user.LastChangePasswordTime)
		if err != nil {
			// Try the standard Casdoor format
			lastChange, err = time.Parse("2006-01-02T15:04:05Z", user.LastChangePasswordTime)
			if err != nil {
				return false // Can't parse, don't expire
			}
		}
	}

	expireAt := lastChange.AddDate(0, 0, organization.PasswordExpireDays)
	return time.Now().After(expireAt)
}
