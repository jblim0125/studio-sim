package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/pkg/errors"

	"github.com/jblim0125/cache-sim/common"
	"github.com/jblim0125/cache-sim/common/appdata"
	"github.com/jblim0125/cache-sim/internal"
	"github.com/jblim0125/cache-sim/internal/stat"
	"github.com/jblim0125/cache-sim/models"
)

// Sender DSL, SID 을 담당하는 객체
type Sender struct {
	log  *common.Logger
	auth *internal.Auth
	dsls *map[string]interface{}
	conf *appdata.Configuration
	stop bool
}

// URL base define
const (
	DSLURL = "http://%s:%d/angora/v2/query/jobs"
	//SIDURL      = "http://%s:%d/angora/v2/query/jobs/%s"
	//SIDCloseURL = "http://%s:%d/angora/v2/query/jobs/%s/close"
)

// NewSender make instance
func (Sender) NewSender(log *common.Logger, auth *internal.Auth,
	dsls *map[string]interface{}, conf *appdata.Configuration) *Sender {
	return &Sender{
		log:  log,
		auth: auth,
		dsls: dsls,
		conf: conf,
		stop: false,
	}
}

// Run start simulator
func (sender *Sender) Run(id int) {
	sender.log.Errorf("[ Sender[ %d ] Start ............................................................... [ OK ]", id)
	sender.log.Infof("Total DSL Count [ %d ]", len(*sender.dsls))
	for {
		sender.DSLLoop(id)
		if sender.stop {
			break
		}
		if !sender.conf.SendRule.Infinite {
			break
		}
	}
	sender.log.Errorf("[ DSL Sender[ %d ] Finish .......................................................... [ OK ]", id)
}

// DSLLoop loop of DSLs
func (sender *Sender) DSLLoop(id int) {
	dslIdx := 1
	for k, v := range *sender.dsls {
		select {
		case <-time.After(time.Duration(sender.conf.SendRule.Period) * time.Millisecond):
			if sender.conf.SendRule.NumThread > 1 {
				if dslIdx%sender.conf.SendRule.NumThread == id {
					sender.log.Debugf("ID[ %02d ] Send[ %02d ] DSL[ %s ]",
						id, sender.conf.SendRule.NumSend, k)
					for i := 0; i < sender.conf.SendRule.NumSend; i++ {
						// 보내기 전에 설정된 최대 DSL 수를 확인
						sender.WaitTotalDSLLimit()
						go sender.RunOneCycle(k, v)
						time.Sleep(time.Duration(sender.conf.SendRule.PeriodDSL) * time.Millisecond)
						if sender.stop {
							break
						}
					}
				}
			}
			dslIdx++
		}
		if sender.stop {
			break
		}
	}
}

// RunOneCycle run cache server one cycle( DSL, SID )
func (sender *Sender) RunOneCycle(k string, v interface{}) {
	RunningLimit{}.IncDSL()
	defer RunningLimit{}.DecDSL()
	// Auth Header
	token, err := sender.auth.GetAuthToken()
	if err != nil {
		sender.log.Errorf("Fail Get Auth Token In Send DSL[ %s ]", err.Error())
		stat.SimStat{}.SendDSLErr()
		return
	}
	_, err = sender.GetSID(token, k, v)
	if err != nil {
		return
	}
	sender.log.Debugf("[ SIM << SERVER ] Receive SID")
}

// GetSID Send DSL And Receive SID
func (sender *Sender) GetSID(token, k string, v interface{}) (string, error) {
	body, err := sender.CrtDSLBody(k, v)
	if err != nil {
		stat.SimStat{}.SendDSLErr()
		return "", errors.Wrap(err, "Fail Create Body From DSL")
	}

	dslURL := fmt.Sprintf(DSLURL, sender.conf.Server.IP, sender.conf.Server.Port)
	req, err := http.NewRequest(http.MethodPost, dslURL, bytes.NewReader(body))
	if err != nil {
		stat.SimStat{}.SendDSLErr()
		return "", errors.Wrap(err, "Fail Create HTTP Request For DSL")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Angora "+token)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		stat.SimStat{}.SendDSLErr()
		return "", errors.Wrap(err, "Fail Send HTTP Request For DSL")
	}
	stat.SimStat{}.SendDSL()
	sender.log.Debugf("[ SIM >> SERVER ] DSL Request")

	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		res := models.DSLResponseBody{}
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&res)
		if err != nil {
			stat.SimStat{}.ErrDSL()
			return "", errors.Wrap(err, "Failed To Decode DSL Response Msg")
		}
		if len(res.SID) <= 0 {
			sender.log.Debugf("Response Code[ %s ]", res.Code)
			return "", fmt.Errorf("Receive ErrCode?[ %s ]", res.Code)
		}
		stat.SimStat{}.RecvSID()
		return res.SID, nil
	default:
		stat.SimStat{}.ErrDSL()
		return "", fmt.Errorf("fail DSL request with status code[ %d ]", resp.StatusCode)
	}
}

// CrtDSLBody DSL(enc, plain) to byte array
func (sender *Sender) CrtDSLBody(k string, v interface{}) ([]byte, error) {
	var err error
	var body []byte
	var encBody models.DSLEncryptRequest
	var plainBody models.DSLPlainRequest

	if sender.conf.SendRule.Encrypt {
		q := []string{}
		s := reflect.ValueOf(v)
		if s.Kind() == reflect.Slice {
			for i := 0; i < s.Len(); i++ {
				q = append(q, s.Index(i).Interface().(string))
			}
		} else {
			q = append(q, s.String())
		}
		encBody = models.DSLEncryptRequest{
			Query:     q,
			Encrypted: false,
		}
		body, err = json.Marshal(encBody)
		if err != nil {
			stat.SimStat{}.SendDSLErr()
			return nil, err
		}
	} else {
		plainBody = models.DSLPlainRequest{
			Query:     k,
			Encrypted: false,
		}
		body, err = json.Marshal(plainBody)
		if err != nil {
			stat.SimStat{}.SendDSLErr()
			return nil, err
		}
	}
	return body, nil
}

// WaitTotalDSLLimit 현재 진행형 상태의 DSL 수를 제한두기 위함.
func (sender *Sender) WaitTotalDSLLimit() {
	now := int32(sender.conf.SendRule.RunningDSL)
	for {
		cnt := RunningLimit{}.GetRunningCnt()
		if cnt < now {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// Destroy 스레드들 종료
func (sender *Sender) Destroy() {
	sender.stop = true
}
