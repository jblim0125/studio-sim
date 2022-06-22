package worker

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jblim0125/cache-sim/internal"
	"github.com/jblim0125/cache-sim/tools/util"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/jblim0125/cache-sim/common"
	"github.com/jblim0125/cache-sim/common/appdata"
	"github.com/jblim0125/cache-sim/internal/stat"
	"github.com/jblim0125/cache-sim/models"
)

// Sender DSL, SID 을 담당하는 객체
type Sender struct {
	log         *common.Logger
	conf        *appdata.Configuration
	auth        *internal.Auth
	dslPath     string
	stop        bool
	destroyWait *sync.WaitGroup
	//dsls map[string]interface{}
	//runningSID  map[string]int
}

// URL base define
const (
	DSLURL      = "http://%s:%d/angora/v2/query/jobs"
	SIDURL      = "http://%s:%d/angora/v2/query/jobs/%s"
	SIDCloseURL = "http://%s:%d/angora/v2/query/jobs/%s/close"
)

// NewSender make instance
func (Sender) NewSender(log *common.Logger, conf *appdata.Configuration,
	auth *internal.Auth, dslPath string) *Sender {
	return &Sender{
		log:         log,
		conf:        conf,
		auth:        auth,
		dslPath:     dslPath,
		stop:        false,
		destroyWait: new(sync.WaitGroup),
	}
}

// Run start simulator
func (sender *Sender) Run(id int) {
	sender.log.Errorf("[ Sender[ %d ] Start ............................................................... [ OK ]", id)
	sender.destroyWait.Add(1)
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
	sender.destroyWait.Done()
}

// DSLLoop loop of DSLs
func (sender *Sender) DSLLoop(id int) {
	dslIdx := 1
	readFile, err := os.Open(sender.dslPath)
	if err != nil {
		sender.log.Errorf("can't open dsl file[ %s ]", err.Error())
		return
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		dsl := fileScanner.Text()

		if dslIdx == 1 {
			if dslIdx%sender.conf.SendRule.NumThread == id {
				sender.log.Debugf("ID[ %02d ] Send[ %02d ] DSL[ %s ]",
					id, sender.conf.SendRule.NumSend, dsl)
				for i := 0; i < sender.conf.SendRule.NumSend; i++ {
					if i == 0 {
						// 보내기 전에 설정된 최대 DSL 수를 확인
						sender.WaitTotalDSLLimit()
						go sender.RunOneCycle(dsl)
					} else {
						select {
						case <-time.After(time.Duration(sender.conf.SendRule.PeriodDSL) * time.Millisecond):
							// 보내기 전에 설정된 최대 DSL 수를 확인
							sender.WaitTotalDSLLimit()
							go sender.RunOneCycle(dsl)
						}
					}
					if sender.stop {
						break
					}
				}
			}
		} else {
			if dslIdx%sender.conf.SendRule.NumThread == id {
				select {
				case <-time.After(time.Duration(sender.conf.SendRule.Period) * time.Millisecond):
					sender.log.Debugf("ID[ %02d ] Send[ %02d ] DSL[ %s ]",
						id, sender.conf.SendRule.NumSend, dsl)
					for i := 0; i < sender.conf.SendRule.NumSend; i++ {
						// 보내기 전에 설정된 최대 DSL 수를 확인
						sender.WaitTotalDSLLimit()
						go sender.RunOneCycle(dsl)
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
	readFile.Close()
}

// RunOneCycle run cache server one cycle( DSL, SID )
func (sender *Sender) RunOneCycle(sendData string) {
	RunningLimit{}.IncDSL()
	defer RunningLimit{}.DecDSL()

	// DSL
	startTime := util.GetMillis()
	stat.SimStat{}.IncDSL()
	sid, err := sender.GetSID(sendData)
	if err != nil {
		sender.log.Errorf("Fail Get SID [ %s ]", err.Error())
		stat.SimStat{}.DecDSL()
		return
	}
	stat.SimStat{}.DecDSL()
	if strings.Index(sid, "cache") >= 0 {
		stat.SimStat{}.ProcInCacheServer()
	} else {
		stat.SimStat{}.ProcInAngora()
	}

	// SID
	stat.SimStat{}.IncSID()
	if err := sender.GetData(sid); err != nil {
		sender.log.Errorf("Fail Get Data [ %s ]", err.Error())
		endTime := util.GetMillis()
		if strings.Index(sid, "cache") >= 0 {
			stat.SimStat{}.SetCacheServerDuration(endTime - startTime)
		} else {
			stat.SimStat{}.SetAngoraDuration(endTime - startTime)
		}
		stat.SimStat{}.DecSID()
		return
	}
	stat.SimStat{}.DecSID()
	endTime := util.GetMillis()
	if strings.Index(sid, "cache") >= 0 {
		stat.SimStat{}.SetCacheServerDuration(endTime - startTime)
	} else {
		stat.SimStat{}.SetAngoraDuration(endTime - startTime)
	}
}

// GetSID Send DSL And Receive SID
func (sender *Sender) GetSID(dsl string) (string, error) {
	body, err := sender.CrtDSLBody(dsl)
	if err != nil {
		stat.SimStat{}.SendDSLErr()
		return "", errors.Wrap(err, "Fail Create Body From DSL")
	}
	token, err := sender.auth.GetAuthToken()
	if err != nil {
		return "", errors.Wrap(err, "failed to get auth token in send close dsl")
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
		stat.SimStat{}.DSLReceiveError()
		return "", errors.Wrap(err, "Failed To Get DSL Response Msg")
	}
	sender.log.Debugf("DSL Response: %s", resBody)

	switch resp.StatusCode {
	case http.StatusOK:
		res := models.ResponseBody{}
		err = json.Unmarshal(resBody, &res)
		if err != nil {
			stat.SimStat{}.DSLReceiveError()
			return "", errors.Wrap(err, "Failed To Unmarshal DSL Response Msg")
		}
		if len(res.SID) <= 0 {
			stat.SimStat{}.DSLReceiveError()
			return "", fmt.Errorf("Receive ErrCode?[ %s ]", res.Code)
		}
		sender.log.Info("[ SIM << SERVER ] Receive SID")
		stat.SimStat{}.DSLReceiveSuccess()
		return res.SID, nil
	default:
		stat.SimStat{}.DSLReceiveError()
		return "", fmt.Errorf("fail DSL request with status code[ %d ]", resp.StatusCode)
	}
}

// CrtDSLBody DSL(enc, plain) to byte array
func (sender *Sender) CrtDSLBody(dsl string) ([]byte, error) {
	var err error
	var body []byte
	var encBody models.DSLEncryptRequest
	var plainBody models.DSLPlainRequest

	if sender.conf.SendRule.Encrypt {
		var q []string
		// TODO : Enc
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
			Query:     dsl,
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
func (sender *Sender) GetData(sid string) error {
	stat.SimStat{}.SendSID()

	token, err := sender.auth.GetAuthToken()
	if err != nil {
		return errors.Wrap(err, "failed to get auth token in send close dsl")
	}

	url := fmt.Sprintf(SIDURL, sender.conf.Server.IP, sender.conf.Server.Port, sid)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		stat.SimStat{}.SendSIDErr()
		return errors.Wrap(err, "failed to create http request for sid request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Angora "+token)

	sender.log.Info("[ SIM >> SERVER ] SID Request")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		stat.SimStat{}.SendSIDErr()
		return errors.Wrap(err, "failed to send sid request")
	}
	//sender.runningSID[sid] = 1
	//defer delete(sender.runningSID, sid)

	sender.log.Info("[ SIM << SERVER ] SID Response")
	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		stat.SimStat{}.SIDReceiveError()
		return errors.Wrap(err, "failed to read sid response")
	}
	sender.log.Debugf("SID Response: %s", resBody)

	switch resp.StatusCode {
	case http.StatusOK:
		res := models.ResponseBody{}
		err = json.Unmarshal(resBody, &res)
		if err != nil {
			stat.SimStat{}.SIDReceiveError()
			return errors.Wrap(err, "Failed To Unmarshal SID Response Msg")
		}
		if len(res.Code) > 0 || len(res.Message) > 0 {
			stat.SimStat{}.SIDReceiveError()
			return fmt.Errorf("Receive ErrCode?[ %s ] IN SID", res.Code)
		}
		sender.log.Info("[ SIM << SERVER ] Receive Data")
		stat.SimStat{}.SIDReceiveSuccess()
		return nil
	default:
		sender.log.Info("[ SIM << SERVER ] Receive Error")
		stat.SimStat{}.SIDReceiveError()
		return fmt.Errorf("fail DSL request with status code[ %d ]", resp.StatusCode)
	}
}

// SendClose send sid close
func (sender *Sender) SendClose() error {
	//var token string
	//var err error
	//
	//// Get Header
	//token, err = sender.auth.GetAuthToken()
	//if err != nil {
	//    return errors.Wrap(err, "failed to get auth token in send close dsl")
	//}
	//
	//client := http.Client{}
	//var req *http.Request
	//var res *http.Response
	//
	//for k := range sender.runningSID {
	//    url := fmt.Sprintf(SIDCloseURL, sender.conf.Server.IP, sender.conf.Server.Port, k)
	//    req, err = http.NewRequest(http.MethodDelete, url, nil)
	//    if err != nil {
	//        sender.log.Errorf("failed to create http request in send close dsl[ %s ]", err.Error())
	//        continue
	//    }
	//    req.Header.Set("Content-Type", "application/json")
	//    req.Header.Set("Authorization", "Angora "+token)
	//
	//    res, err = client.Do(req)
	//    if err != nil {
	//        sender.log.Errorf("failed to request DSL CLOSE[ %s ]", err.Error())
	//        continue
	//    }
	//    res.Body.Close()
	//}
	return nil
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

// DestroyWait 스레드들 종료
func (sender *Sender) DestroyWait() {
	sender.destroyWait.Wait()
	sender.SendClose()
}
