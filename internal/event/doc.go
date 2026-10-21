// Package event holds the messages the domains exchange through the broker: their names, destinations and
// payloads. A message connects the domain that saves it to the domains that handle it, usually in another
// deployable, so it belongs to neither of them: domains import event, never each other.
package event
