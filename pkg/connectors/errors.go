// Package connectors provides connectors to storage/databases/etc
package connectors

import "errors"

var (
	// ErrPostgresConnectorNotInitialized postgres connector not initialized
	ErrPostgresConnectorNotInitialized = errors.New("postgres connector not initialized")
)
