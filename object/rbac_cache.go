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
	"sync"
	"time"
)

// rbacCacheEntry stores cached role/permission data with TTL.
type rbacCacheEntry struct {
	roles     []*Role
	expireAt  time.Time
}

var (
	rbacCache     = sync.Map{}
	rbacCacheTTL  = 60 * time.Second
)

// getUserRolesWithCache returns user roles from cache or DB.
// Cache entries expire after 60 seconds.
func getUserRolesWithCache(userId string) ([]*Role, error) {
	if entry, ok := rbacCache.Load(userId); ok {
		cached := entry.(*rbacCacheEntry)
		if time.Now().Before(cached.expireAt) {
			return cached.roles, nil
		}
		rbacCache.Delete(userId)
	}

	// Cache miss or expired — fetch from DB
	roles, err := getRolesByUser(userId)
	if err != nil {
		return nil, err
	}

	rbacCache.Store(userId, &rbacCacheEntry{
		roles:    roles,
		expireAt: time.Now().Add(rbacCacheTTL),
	})

	return roles, nil
}

// InvalidateRbacCache removes a user's cached RBAC data.
// Call this when roles or permissions are modified.
func InvalidateRbacCache(userId string) {
	rbacCache.Delete(userId)
}

// InvalidateAllRbacCache clears the entire RBAC cache.
// Call this when a role/permission definition changes (affects all users).
func InvalidateAllRbacCache() {
	rbacCache.Range(func(key, value interface{}) bool {
		rbacCache.Delete(key)
		return true
	})
}
