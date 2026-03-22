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
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

// ApiKey represents a persistent API key for machine-to-machine authentication.
type ApiKey struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	KeyPrefix   string `xorm:"varchar(20)" json:"keyPrefix"`   // First 8 chars of raw key (for identification)
	KeyHash     string `xorm:"varchar(100)" json:"keyHash"`     // SHA-256 hash of full key
	Scopes      string `xorm:"varchar(1000)" json:"scopes"`     // Comma-separated scope list
	ExpiresAt   string `xorm:"varchar(100)" json:"expiresAt"`   // Empty = never expires
	LastUsedAt  string `xorm:"varchar(100)" json:"lastUsedAt"`
	IsEnabled   bool   `json:"isEnabled"`
}

const apiKeyPrefix = "casdoor_ak_"

func (ak *ApiKey) GetId() string {
	return fmt.Sprintf("%s/%s", ak.Owner, ak.Name)
}

// GenerateApiKey creates a new API key and returns the raw key (shown only once).
// The raw key format is: casdoor_ak_<32 random hex chars>
func GenerateApiKey() (rawKey string, prefix string, hash string, err error) {
	bytes := make([]byte, 32)
	_, err = rand.Read(bytes)
	if err != nil {
		return "", "", "", err
	}

	rawKey = apiKeyPrefix + hex.EncodeToString(bytes)
	prefix = rawKey[:len(apiKeyPrefix)+8]

	h := sha256.Sum256([]byte(rawKey))
	hash = hex.EncodeToString(h[:])

	return rawKey, prefix, hash, nil
}

// ValidateApiKey checks if a raw API key is valid and returns the associated ApiKey record.
func ValidateApiKey(rawKey string) (*ApiKey, error) {
	h := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(h[:])

	apiKey := &ApiKey{}
	existed, err := ormer.Engine.Where("key_hash = ? AND is_enabled = ?", keyHash, true).Get(apiKey)
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, nil
	}

	// Check expiration
	if apiKey.ExpiresAt != "" {
		// Simple string comparison works for ISO 8601 format
		now := util.GetCurrentTime()
		if now > apiKey.ExpiresAt {
			return nil, nil
		}
	}

	// Update last used timestamp
	apiKey.LastUsedAt = util.GetCurrentTime()
	_, _ = ormer.Engine.ID(core.PK{apiKey.Owner, apiKey.Name}).Cols("last_used_at").Update(apiKey)

	return apiKey, nil
}

func GetApiKeys(owner string) ([]*ApiKey, error) {
	apiKeys := []*ApiKey{}
	err := ormer.Engine.Desc("created_time").Find(&apiKeys, &ApiKey{Owner: owner})
	return apiKeys, err
}

func GetApiKey(owner, name string) (*ApiKey, error) {
	apiKey := &ApiKey{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(apiKey)
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, nil
	}
	return apiKey, nil
}

func AddApiKey(apiKey *ApiKey) (bool, string, error) {
	rawKey, prefix, hash, err := GenerateApiKey()
	if err != nil {
		return false, "", err
	}

	apiKey.CreatedTime = util.GetCurrentTime()
	apiKey.KeyPrefix = prefix
	apiKey.KeyHash = hash
	apiKey.IsEnabled = true

	affected, err := ormer.Engine.Insert(apiKey)
	if err != nil {
		return false, "", err
	}

	return affected != 0, rawKey, nil
}

func DeleteApiKey(owner, name string) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{owner, name}).Delete(&ApiKey{})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}
