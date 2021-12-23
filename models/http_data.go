package models

import "net/http"

// HTTPData non blocking을 위한 채널 통신 데이터
type HTTPData struct {
	Response *http.Response
	Error    error
}
