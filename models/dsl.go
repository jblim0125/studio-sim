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

// ResponseBody angora/query/job, angora/v2/query/job response body
type ResponseBody struct {
	SID     string `json:"sid,omitempty"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}
