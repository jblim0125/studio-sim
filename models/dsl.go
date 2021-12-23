package models

// DSLPlainRequest Request POST angora/query/job, angora/v2/query/job request body
type DSLPlainRequest struct {
	Query     string `json:"q"`
	Encrypted bool   `json:"is_secret"`
}

// DSLEncryptRequest Request POST angora/query/job, angora/v2/query/job request body
type DSLEncryptRequest struct {
	Query     []string `json:"q"`
	Encrypted bool     `json:"is_secret"`
}

// DSLResponseBody angora/query/job, angora/v2/query/job response body
type DSLResponseBody struct {
	SID string `json:"sid"`
}
