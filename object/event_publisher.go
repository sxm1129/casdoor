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

// EventPublisher defines the interface for publishing events to various backends.
// This abstraction allows switching between Webhook, Kafka, Redis Pub/Sub, etc.
type EventPublisher interface {
	// Publish sends an event to the configured backend.
	Publish(eventType string, payload interface{}) error

	// Name returns the publisher name for logging/debugging.
	Name() string
}

// WebhookEventPublisher publishes events via the webhook delivery system.
type WebhookEventPublisher struct{}

func (p *WebhookEventPublisher) Publish(eventType string, payload interface{}) error {
	// Currently handled by the existing webhook delivery system.
	// Future: direct webhook enqueue based on event type.
	return nil
}

func (p *WebhookEventPublisher) Name() string {
	return "webhook"
}

// eventPublishers holds registered event publishers.
var eventPublishers []EventPublisher

// RegisterEventPublisher adds a new event publisher to the registry.
func RegisterEventPublisher(publisher EventPublisher) {
	eventPublishers = append(eventPublishers, publisher)
}

// PublishEvent sends an event to all registered publishers.
func PublishEvent(eventType string, payload interface{}) {
	for _, publisher := range eventPublishers {
		if err := publisher.Publish(eventType, payload); err != nil {
			// Log but don't fail — event publishing should not block business logic
			_ = err
		}
	}
}
