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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/casdoor/casdoor/util"
)

// WebhookDelivery tracks each individual webhook delivery attempt.
type WebhookDelivery struct {
	Id          int    `xorm:"int notnull pk autoincr" json:"id"`
	WebhookId   string `xorm:"varchar(200) index" json:"webhookId"`
	RecordId    int    `xorm:"int index" json:"recordId"`
	DeliveryId  string `xorm:"varchar(100) unique" json:"deliveryId"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Status       string `xorm:"varchar(20) index" json:"status"` // pending, success, failed
	Attempts     int    `xorm:"int" json:"attempts"`
	MaxAttempts  int    `xorm:"int" json:"maxAttempts"`
	LastAttemptAt string `xorm:"varchar(100)" json:"lastAttemptAt"`
	NextRetryAt  string `xorm:"varchar(100)" json:"nextRetryAt"`

	Url          string `xorm:"varchar(500)" json:"url"`
	Method       string `xorm:"varchar(20)" json:"method"`
	RequestBody  string `xorm:"mediumtext" json:"requestBody"`
	Signature    string `xorm:"varchar(200)" json:"signature"`

	ResponseCode int    `xorm:"int" json:"responseCode"`
	ResponseBody string `xorm:"text" json:"responseBody"`
	ErrorMessage string `xorm:"text" json:"errorMessage"`
}

const (
	DeliveryStatusPending = "pending"
	DeliveryStatusSuccess = "success"
	DeliveryStatusFailed  = "failed"

	DefaultMaxAttempts = 3
)

func AddWebhookDelivery(delivery *WebhookDelivery) (int64, error) {
	delivery.DeliveryId = util.GenerateId()
	delivery.CreatedTime = util.GetCurrentTime()
	delivery.Status = DeliveryStatusPending
	delivery.Attempts = 0
	delivery.MaxAttempts = DefaultMaxAttempts
	return ormer.Engine.Insert(delivery)
}

func GetPendingWebhookDeliveries() ([]*WebhookDelivery, error) {
	deliveries := []*WebhookDelivery{}
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	err := ormer.Engine.Where("status = ? AND (next_retry_at = '' OR next_retry_at <= ?)", DeliveryStatusPending, now).
		OrderBy("created_time ASC").
		Limit(50).
		Find(&deliveries)
	return deliveries, err
}

func UpdateWebhookDelivery(delivery *WebhookDelivery) error {
	_, err := ormer.Engine.ID(delivery.Id).AllCols().Update(delivery)
	return err
}

func GetWebhookDeliveries(webhookId string, offset, limit int) ([]*WebhookDelivery, error) {
	deliveries := []*WebhookDelivery{}
	session := ormer.Engine.Desc("created_time")
	if webhookId != "" {
		session = session.Where("webhook_id = ?", webhookId)
	}
	if offset >= 0 && limit > 0 {
		session = session.Limit(limit, offset)
	}
	err := session.Find(&deliveries)
	return deliveries, err
}

func GetWebhookDeliveryCount(webhookId string) (int64, error) {
	delivery := &WebhookDelivery{}
	if webhookId != "" {
		delivery.WebhookId = webhookId
	}
	return ormer.Engine.Count(delivery)
}

// EnqueueWebhookDeliveries creates WebhookDelivery records for each matching webhook.
// The background worker will pick up and send them asynchronously.
func EnqueueWebhookDeliveries(record *Record) error {
	webhooks, err := getWebhooksByOrganization("")
	if err != nil {
		return err
	}

	webhooks = getFilteredWebhooks(webhooks, record.Organization, record.Action)

	for _, webhook := range webhooks {
		record2 := *record
		if len(webhook.ObjectFields) != 0 && webhook.ObjectFields[0] != "All" {
			record2.Object = filterRecordObject(record.Object, webhook.ObjectFields)
		}

		body := util.StructToJson(&record2)

		delivery := &WebhookDelivery{
			WebhookId:   webhook.GetId(),
			RecordId:    record.Id,
			Url:         webhook.Url,
			Method:      webhook.Method,
			RequestBody: body,
		}

		if _, err := AddWebhookDelivery(delivery); err != nil {
			return fmt.Errorf("enqueue delivery for webhook %s: %v", webhook.GetId(), err)
		}
	}

	return nil
}

// ComputeWebhookSignature generates an HMAC-SHA256 signature for webhook payload verification.
func ComputeWebhookSignature(payload string, secret string) string {
	if secret == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return fmt.Sprintf("sha256=%s", hex.EncodeToString(mac.Sum(nil)))
}
