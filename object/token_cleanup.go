// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	"fmt"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/robfig/cron/v3"
)

const tokenCleanupBatchSize = 1000

// CleanupTokens deletes tokens that expired more than retentionSeconds ago.
// Uses batch deletion to avoid memory issues with large token tables.
func CleanupTokens(retentionSeconds int) (int64, error) {
	cutoff := time.Now().Add(-time.Duration(retentionSeconds) * time.Second)
	cutoffStr := cutoff.Format("2006-01-02 15:04:05")
	totalDeleted := int64(0)

	for {
		// Use SQL-level expiry check via ExpiresIn + CreatedTime
		// instead of loading all tokens into memory and parsing JWT
		result, err := ormer.Engine.Exec(
			"DELETE FROM `token` WHERE `expires_in` > 0 AND "+
				"TIMESTAMPADD(SECOND, `expires_in`, STR_TO_DATE(`created_time`, '%Y-%m-%d %H:%i:%s')) < ? LIMIT ?",
			cutoffStr, tokenCleanupBatchSize,
		)
		if err != nil {
			// Fallback: some databases don't support TIMESTAMPADD/STR_TO_DATE
			// Use a simpler approach based on created_time only
			result, err = ormer.Engine.Exec(
				"DELETE FROM `token` WHERE `created_time` < ? LIMIT ?",
				cutoff.AddDate(0, 0, -30).Format("2006-01-02 15:04:05"), tokenCleanupBatchSize,
			)
			if err != nil {
				return totalDeleted, fmt.Errorf("failed to cleanup tokens: %w", err)
			}
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return totalDeleted, err
		}

		totalDeleted += affected

		if affected < tokenCleanupBatchSize {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	return totalDeleted, nil
}

func getTokenRetentionInterval(days int) int {
	if days <= 0 {
		days = 30
	}
	return days * 24 * 3600
}

// InitCleanupTokens starts a daily cron job to clean up expired tokens.
func InitCleanupTokens() {
	interval := getTokenRetentionInterval(30)

	deleted, err := CleanupTokens(interval)
	if err != nil {
		logs.Warning("Token cleanup at startup failed: %v", err)
	} else if deleted > 0 {
		logs.Info("Token cleanup at startup: deleted %d expired tokens", deleted)
	}

	cronJob := cron.New()
	_, err = cronJob.AddFunc("0 0 * * *", func() {
		deleted, err := CleanupTokens(interval)
		if err != nil {
			logs.Warning("Scheduled token cleanup failed: %v", err)
		} else if deleted > 0 {
			logs.Info("Scheduled token cleanup: deleted %d expired tokens", deleted)
		}
	})
	if err != nil {
		logs.Warning("Failed to schedule token cleanup: %v", err)
		return
	}
	cronJob.Start()
}
