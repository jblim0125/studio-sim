package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	log        *common.Logger
	auth       *internal.Auth
	dsls       *map[string]interface{}
	conf       *appdata.Configuration
	runningSID map[string]int
	stop       bool
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
		log:        log,
		auth:       auth,
		dsls:       dsls,
		conf:       conf,
		stop:       false,
		runningSID: make(map[string]int),
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
		if dslIdx == 1 {
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
		} else {
			select {
			case <-time.After(time.Duration(sender.conf.SendRule.Period) * time.Millisecond):
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
		}
		dslIdx++
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
	sid, err := sender.GetSID(token, k, v)
	if err != nil {
		sender.log.Errorf("Fail Get SID [ %s ]", err.Error())
		return
	}
	if err := sender.GetData(token, sid); err != nil {
		sender.log.Errorf("Fail Get Data [ %s ]", err.Error())
		return
	}
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

	// Send
	stat.SimStat{}.SendDSL()
	sender.log.Info("[ SIM >> SERVER ] DSL Request")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		stat.SimStat{}.SendDSLErr()
		return "", errors.Wrap(err, "Fail Send HTTP Request For DSL")
	}

	sender.log.Info("[ SIM << SERVER ] DSL Response")
	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		stat.SimStat{}.ErrDSL()
		return "", errors.Wrap(err, "Failed To Get DSL Response Msg")
	}
	sender.log.Debugf("DSL Response: %s", resBody)

	switch resp.StatusCode {
	case http.StatusOK:
		res := models.ResponseBody{}
		err = json.Unmarshal(resBody, &res)
		if err != nil {
			stat.SimStat{}.ErrDSL()
			return "", errors.Wrap(err, "Failed To Unmarshal DSL Response Msg")
		}
		if len(res.SID) <= 0 {
			stat.SimStat{}.ErrDSL()
			return "", fmt.Errorf("Receive ErrCode?[ %s ]", res.Code)
		}
		sender.log.Info("[ SIM << SERVER ] Receive SID")
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

// GetData SID 를 이용해 데이터 요청
func (sender *Sender) GetData(token, sid string) error {
	client := http.Client{}
	var req *http.Request
	var err error

	url := fmt.Sprintf(SIDURL, sender.conf.Server.IP, sender.conf.Server.Port, sid)
	req, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		stat.SimStat{}.SendSIDErr()
		return errors.Wrap(err, "failed to create http request for sid request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Angora "+token)

	sender.log.Info("[ SIM >> SERVER ] SID Request")
	stat.SimStat{}.SendSID()
	sender.runningSID[sid] = 1
	defer delete(sender.runningSID, sid)
	resp, err := client.Do(req)
	if err != nil {
		stat.SimStat{}.SendSIDErr()
		return errors.Wrap(err, "failed to send sid request")
	}
	sender.log.Info("[ SIM << SERVER ] SID Response")
	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		stat.SimStat{}.ErrSID()
		return errors.Wrap(err, "failed to read sid response")
	}
	sender.log.Debugf("SID Response: %s", resBody)

	switch resp.StatusCode {
	case http.StatusOK:
		res := models.ResponseBody{}
		err = json.Unmarshal(resBody, &res)
		if err != nil {
			stat.SimStat{}.ErrSID()
			return errors.Wrap(err, "Failed To Unmarshal SID Response Msg")
		}
		if len(res.Code) > 0 || len(res.Message) > 0 {
			stat.SimStat{}.ErrSID()
			return fmt.Errorf("Receive ErrCode?[ %s ] IN SID", res.Code)
		}
		sender.log.Info("[ SIM << SERVER ] Receive Data")
		stat.SimStat{}.RecvData()
		return nil
	default:
		stat.SimStat{}.ErrSID()
		return fmt.Errorf("fail DSL request with status code[ %d ]", resp.StatusCode)
	}
}

//// SendClose send sid close
//func (sidSender *SIDSender) SendClose() {
//    var token string
//    var err error
//
//    // Get Header
//    token, err = sidSender.Auth.GetAuthToken()
//    if err != nil {
//        sidSender.log.Errorf("Fail Get Auth Token In Send DSL[ %s ]", err.Error())
//        return
//    }
//
//    client := http.Client{}
//    var req *http.Request
//    var res *http.Response
//
//    for k := range sidSender.RunningSID {
//
//        url := fmt.Sprintf(SIDURL, sidSender.Conf.Server.IP, sidSender.Conf.Server.Port, k)
//        req, err = http.NewRequest(http.MethodDelete, url, nil)
//        if err != nil {
//            sidSender.log.Errorf("Failed To Create Http.Request[ %s ]", err.Error())
//            return
//        }
//        req.Header.Set("Content-Type", "application/json")
//        req.Header.Set("Authorization", "Angora "+token)
//
//        res, err = client.Do(req)
//        if err != nil {
//            sidSender.log.Errorf("Failed To Request JOB CLOSE[ %s ]", err.Error())
//            return
//        }
//        res.Body.Close()
//    }
//}

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
