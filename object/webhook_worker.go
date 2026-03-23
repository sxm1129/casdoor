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

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/util"
)

// StartWebhookWorker launches a background goroutine that processes pending webhook deliveries.
// It runs every 5 seconds, picks up pending deliveries, and attempts to send them
// with exponential backoff retry (3 attempts: 0s, 4s, 16s delays).
func StartWebhookWorker() {
	util.SafeGoroutine(func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			processWebhookDeliveries()
		}
	})
}

func processWebhookDeliveries() {
	deliveries, err := GetPendingWebhookDeliveries()
	if err != nil {
		logs.Error("Failed to get pending webhook deliveries: %s", err)
		return
	}

	for _, delivery := range deliveries {
		processOneDelivery(delivery)
	}
}

func processOneDelivery(delivery *WebhookDelivery) {
	delivery.Attempts++
	delivery.LastAttemptAt = util.GetCurrentTime()

	// Parse the webhook to get headers etc.
	webhook, err := GetWebhook(delivery.WebhookId)
	if err != nil || webhook == nil {
		delivery.Status = DeliveryStatusFailed
		delivery.ErrorMessage = "webhook not found or deleted"
		_ = UpdateWebhookDelivery(delivery)
		return
	}

	// Compute signature
	if delivery.Signature == "" {
		delivery.Signature = ComputeWebhookSignature(delivery.RequestBody, webhook.Name)
	}

	// Attempt HTTP delivery
	statusCode, respBody, sendErr := sendWebhookHTTP(webhook, delivery)

	delivery.ResponseCode = statusCode
	if len(respBody) > 500 {
		respBody = respBody[:500]
	}
	delivery.ResponseBody = respBody

	if sendErr != nil {
		delivery.ErrorMessage = sendErr.Error()
	}

	// Evaluate result
	if statusCode >= 200 && statusCode < 300 && sendErr == nil {
		delivery.Status = DeliveryStatusSuccess
		WebhookDeliveries.WithLabelValues("success").Inc()
	} else if delivery.Attempts >= delivery.MaxAttempts {
		delivery.Status = DeliveryStatusFailed
		WebhookDeliveries.WithLabelValues("failed").Inc()
	} else {
		// Schedule retry with exponential backoff: 4^(attempt-1) seconds
		backoff := time.Duration(1<<(2*delivery.Attempts)) * time.Second
		if backoff > 5*time.Minute {
			backoff = 5 * time.Minute
		}
		nextRetry := time.Now().UTC().Add(backoff)
		delivery.NextRetryAt = nextRetry.Format("2006-01-02T15:04:05Z")
	}

	if err := UpdateWebhookDelivery(delivery); err != nil {
		logs.Error("Failed to update webhook delivery %d: %s", delivery.Id, err)
	}
}

// sendWebhookHTTP performs the actual HTTP call for a webhook delivery.
// It reuses the existing sendWebhook function signature and adds delivery-specific headers.
func sendWebhookHTTP(webhook *Webhook, delivery *WebhookDelivery) (int, string, error) {
	// Build a temporary Record from the delivery for the existing sendWebhook function
	record := &Record{
		Object: delivery.RequestBody,
	}

	// We use the existing sendWebhook function which handles HTTP method, headers, content-type etc.
	statusCode, respBody, err := sendWebhook(webhook, record, nil)
	return statusCode, respBody, err
}
