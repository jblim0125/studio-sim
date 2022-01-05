package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jblim0125/studio-sim/common"
	"github.com/jblim0125/studio-sim/common/appdata"
	"github.com/jblim0125/studio-sim/internal"
	"github.com/jblim0125/studio-sim/internal/stat"
	"github.com/jblim0125/studio-sim/models"
	"github.com/jblim0125/studio-sim/tools/util"
	"net/http"
	"reflect"
	"time"
)

// DSLSender DSL 전송 객체
type DSLSender struct {
	log  *common.Logger
	Auth *internal.Auth
	Dsls *map[string]interface{}
	Conf *appdata.Configuration
	STOP bool
}

// DslURL DSL Request 요청 URL
const DslURL = "http://%s:%d/angora/v2/query/jobs"

// NewDSLSender DSL 전송 goroutine 생성
func (DSLSender) NewDSLSender(log *common.Logger, auth *internal.Auth,
	dsls *map[string]interface{}, conf *appdata.Configuration) *DSLSender {
	return &DSLSender{
		log:  log,
		Auth: auth,
		Dsls: dsls,
		Conf: conf,
		STOP: false,
	}
}

// Run 환경 설정에 맞춰
func (dslSender *DSLSender) Run(id int, ch chan models.HTTPData) {
	var now, next, remain int64
	var period int64 = int64(dslSender.Conf.SendRule.Period)
	now = util.GetMillis()
	remain = period - (now % period)
	next = (now + remain)

	dslSender.log.Errorf("[ DSL Sender[ %d ] Start ........................................................... [ OK ]", id)
	url := fmt.Sprintf(DslURL, dslSender.Conf.Server.IP, dslSender.Conf.Server.Port)
	for {
		for k, v := range *dslSender.Dsls {
			now = util.GetMillis()
			if now >= next {
				for i := 0; i < dslSender.Conf.SendRule.NumSend; i++ {
					dslSender.WaitTotalDSLLimit()
					go func() {
						err := dslSender.SendDSL(url, k, v, ch)
						//err := dslSender.TestSendDSL(k, v, ch)
						if err != nil {
							dslSender.log.Errorf("Send DSL Err[ %s ]", err.Error())
						} else {
							RunningDSL{}.IncDSL()
						}
					}()
					time.Sleep(time.Duration(dslSender.Conf.SendRule.PeriodDSL) * time.Millisecond)
				}
				// Calc Next Runtime
				now = util.GetMillis()
				remain = period - (now % period)
				next = (now + remain)
			}
			// Sleep
			time.Sleep(20 * time.Millisecond)
			if dslSender.STOP {
				break
			}
		}
		if dslSender.STOP {
			break
		}
		if !dslSender.Conf.SendRule.Infinite {
			break
		}
	}
	dslSender.log.Errorf("[ DSL Sender[ %d ] Finish .......................................................... [ OK ]", id)
}

// SendDSL DSL 전송
func (dslSender *DSLSender) SendDSL(url, k string, v interface{}, ch chan models.HTTPData) error {
	client := http.Client{}
	var req *http.Request
	var err error
	var body []byte
	var encBody models.DSLEncryptRequest
	var plainBody models.DSLPlainRequest

	// Set Header
	token, err := dslSender.Auth.GetAuthToken()
	if err != nil {
		dslSender.log.Errorf("Fail Get Auth Token In Send DSL[ %s ]", err.Error())
		stat.SimStat{}.SendDSLErr()
		return err
	}

	if dslSender.Conf.SendRule.Encrypt {
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
			return err
		}
	} else {
		plainBody = models.DSLPlainRequest{
			Query:     k,
			Encrypted: false,
		}
		body, err = json.Marshal(plainBody)
		if err != nil {
			stat.SimStat{}.SendDSLErr()
			return err
		}
	}
	req, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		stat.SimStat{}.SendDSLErr()
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Angora "+token)
	resp, err := client.Do(req)
	if err != nil {
		stat.SimStat{}.SendDSLErr()
		return nil
	}
	stat.SimStat{}.SendDSL()
	dslSender.log.Debugf("[ SIM >> SERVER ] DSL Request")

	ch <- models.HTTPData{
		Response: resp,
		Error:    err,
	}
	return nil
}

// TestSendDSL test 용도
func (dslSender *DSLSender) TestSendDSL(k string, v interface{}, ch chan models.HTTPData) error {
	req, err := http.NewRequest(http.MethodGet, "http://211.232.75.75:8090", nil)
	if err != nil {
		return nil
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	stat.SimStat{}.SendDSL()

	ch <- models.HTTPData{
		Response: resp,
		Error:    err,
	}
	return nil

}

// Destroy DSL 전송 스레드들 종료
func (dslSender *DSLSender) Destroy() {
	dslSender.STOP = true
}

// WaitTotalDSLLimit 현재 진행형 상태의 DSL 수를 제한두기 위함.
func (dslSender *DSLSender) WaitTotalDSLLimit() {
	now := int32(dslSender.Conf.SendRule.RunningDSL)
	for {
		cnt := RunningDSL{}.GetDSL()
		if cnt < now {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}
