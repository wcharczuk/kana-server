package controller

import "github.com/blend/go-sdk/uuid"

// UUIDValue returns a uuid typed value.
func UUIDValue(param string, inputErr error) (uuid.UUID, error) {
	if inputErr != nil {
		return nil, inputErr
	}
	return uuid.Parse(param)
}
