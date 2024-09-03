package model

import "github.com/google/uuid"

type Bucket struct {
	Id          uuid.UUID `json:"id"`
	EndpointUrl string    `json:"endpoint_url"`
	Name        string    `json:"name"`
	Active      bool      `json:"active"`
}
