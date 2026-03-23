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
	"fmt"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/robfig/cron/v3"
)

const (
	defaultRecordRetentionDays = 90
	recordCleanupBatchSize     = 1000
)

// CleanupOldRecords deletes audit records older than retentionDays.
// Uses batch deletion to avoid locking large tables.
func CleanupOldRecords(retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		retentionDays = defaultRecordRetentionDays
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays).Format("2006-01-02 15:04:05")
	totalDeleted := int64(0)

	for {
		result, err := ormer.Engine.Exec(
			"DELETE FROM `record` WHERE `created_time` < ? LIMIT ?",
			cutoff, recordCleanupBatchSize,
		)
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to cleanup records: %w", err)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return totalDeleted, err
		}

		totalDeleted += affected

		// If fewer rows deleted than batch size, we're done
		if affected < recordCleanupBatchSize {
			break
		}

		// Brief pause between batches to reduce DB pressure
		time.Sleep(100 * time.Millisecond)
	}

	return totalDeleted, nil
}

// InitRecordCleanup starts a daily cron job to clean up old audit records.
func InitRecordCleanup() {
	// Run once at startup
	deleted, err := CleanupOldRecords(defaultRecordRetentionDays)
	if err != nil {
		logs.Warning("Record cleanup at startup failed: %v", err)
	} else if deleted > 0 {
		logs.Info("Record cleanup at startup: deleted %d old records (>%d days)", deleted, defaultRecordRetentionDays)
	}

	// Schedule daily at 02:00
	cronJob := cron.New()
	_, err = cronJob.AddFunc("0 2 * * *", func() {
		deleted, err := CleanupOldRecords(defaultRecordRetentionDays)
		if err != nil {
			logs.Warning("Scheduled record cleanup failed: %v", err)
		} else if deleted > 0 {
			logs.Info("Scheduled record cleanup: deleted %d old records", deleted)
		}
	})
	if err != nil {
		logs.Warning("Failed to schedule record cleanup: %v", err)
		return
	}
	cronJob.Start()
}
