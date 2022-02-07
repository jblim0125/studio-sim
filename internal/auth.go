package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jblim0125/cache-sim/models"
	"github.com/jblim0125/cache-sim/tools/util"
)

/*
curl -X POST -H "Content-Type: application/json" \
-d " \
{ \
 \"password\":\"b\", \
 \"encrypted\": true,\
 \"group_id\": \"\", \
 \"id\": \"root\" \
} \
" \
http://10.1.118.40:6036/angora/auth

*/

// Auth auth class
type Auth struct {
	URL        string
	Token      string
	ExpireTime int64
}

const authURL string = "http://%s:%d/angora/auth"

// Initialize Get Angora Auth Token
func (Auth) Initialize(ip string, port int) (*Auth, error) {
	auth := &Auth{
		URL: fmt.Sprintf(authURL, ip, port),
	}
	if err := auth.sendAuthRequest(); err != nil {
		return nil, err
	}
	return auth, nil
}

// GetAuthToken get auth token
func (a *Auth) GetAuthToken() (string, error) {
	if a.ExpireTime+(1000*60*5) <= util.GetMillis() {
		if err := a.sendAuthRequest(); err != nil {
			return "", fmt.Errorf("failed to refresh aut token[ %s ]", err.Error())
		}
	}
	return a.Token, nil
}

// sendAuthRequest send auth request
func (a *Auth) sendAuthRequest() error {
	requestBytes, err := json.Marshal(models.AuthReqBody{
		Password:  "perfTest",
		Encrypted: true,
		GroupID:   "",
		ID:        "root",
	})
	if err != nil {
		return fmt.Errorf("failed to marshal auth request[ %s ]", err.Error())
	}
	resp, err := http.Post(a.URL, "application/json", bytes.NewReader(requestBytes))
	if err != nil {
		return fmt.Errorf("auth fail. http post [ %s ]", err.Error())
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		res := models.AuthResBody{}
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&res)
		if err != nil && err != io.EOF {
			return err
		}
		if strings.ToLower(res.Status) != "ok" {
			return fmt.Errorf("auth fail with status[ %s ]", res.Status)
		}
		a.Token = res.Token
		a.ExpireTime = util.GetMillis()
		return nil
	default:
		return fmt.Errorf("auth fail with status code[ %d ]", resp.StatusCode)
	}
}
