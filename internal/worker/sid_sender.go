package worker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jblim0125/cache-sim/common"
	"github.com/jblim0125/cache-sim/common/appdata"
	"github.com/jblim0125/cache-sim/internal"
	"github.com/jblim0125/cache-sim/internal/stat"
	"github.com/jblim0125/cache-sim/models"
)

// SIDSender SID 전송 객체
type SIDSender struct {
	log        *common.Logger
	DslChannel []chan models.HTTPData
	SidChannel []chan string
	Auth       *internal.Auth
	Conf       *appdata.Configuration
	RunningSID map[string]int
	STOP       bool
}

// NewSIDSender SID 전송 goroutine 생성
func (SIDSender) NewSIDSender(log *common.Logger,
	auth *internal.Auth, conf *appdata.Configuration) *SIDSender {
	sidSender := &SIDSender{
		log:        log,
		Auth:       auth,
		Conf:       conf,
		RunningSID: make(map[string]int),
		STOP:       false,
	}

	for i := 0; i < conf.SendRule.NumThread; i++ {
		dslCh := make(chan models.HTTPData)
		sidSender.DslChannel = append(sidSender.DslChannel, dslCh)
		sidCh := make(chan string)
		sidSender.SidChannel = append(sidSender.SidChannel, sidCh)
	}
	return sidSender
}

// RunSIDReceiver angora/query/job 의 응답을 수신하고, 결과를 SID 전송 goroutine으로 전달
func (sidSender *SIDSender) RunSIDReceiver(id int) {
	sidSender.log.Errorf("[ DSL Response Receiver [ %d ] Start ............................................... [ OK ]", id)
	for {
		select {
		case resp := <-sidSender.DslChannel[id]:
			sidSender.log.Debugf("[ SIM << SERVER ] DSL Response")
			err := sidSender.ReadDSLResponse(id, resp)
			if err != nil {
				sidSender.log.Errorf("SID Receive Err[ %s ]", err.Error())
				RunningLimit{}.DecDSL()
			}
		}
		if sidSender.STOP {
			break
		}
	}
	sidSender.log.Errorf("[ DSL Response Receiver [ %d ] Stop ................................................ [ OK ]", id)
}

// ReadDSLResponse DSL 응답 메시지 확인
func (sidSender *SIDSender) ReadDSLResponse(id int, resp models.HTTPData) error {
	defer resp.Response.Body.Close()
	switch resp.Response.StatusCode {
	case http.StatusOK:
		res := models.ResponseBody{}
		decoder := json.NewDecoder(resp.Response.Body)
		err := decoder.Decode(&res)
		if err != nil && err != io.EOF {
			stat.SimStat{}.ErrDSL()
			sidSender.log.Errorf("[ SIM << SERVER ] Error DSL Response")
			return err
		}
		sidSender.log.Debugf("[ SIM << SERVER ] Receive SID")
		stat.SimStat{}.RecvSID()
		sidSender.SidChannel[id] <- res.SID
		//sidSender.SidChannel[id] <- "test"
		return nil
	default:
		stat.SimStat{}.ErrDSL()
		return fmt.Errorf("fail DSL request with status code[ %d ]", resp.Response.StatusCode)
	}
}

// RunSIDSender angora/query/job/{sid} 전송
func (sidSender *SIDSender) RunSIDSender(id int) {
	sidSender.log.Errorf("[ SID Request Sender [ %d ] Start .................................................. [ OK ]", id)
	for {
		select {
		case sid := <-sidSender.SidChannel[id]:
			go func() {
				sidSender.log.Debugf("[ SIM >> SERVER ] SID Request")
				err := sidSender.SendSID(sid)
				if err != nil {
					sidSender.log.Errorf("Send SID Err[ %s ]", err.Error())
				}
				RunningLimit{}.DecDSL()
			}()
		}
		if sidSender.STOP {
			break
		}
	}
	sidSender.log.Errorf("[ SID Request Sender [ %d ] Stop ................................................... [ OK ]", id)
}

// Destroy SID Receiver, SID Sender 종료
func (sidSender *SIDSender) Destroy() {
	sidSender.STOP = true
	if len(sidSender.RunningSID) > 0 {
		sidSender.log.Errorf("SID Sender Destroy Need Close Msg")
		sidSender.SendClose()
	}
}

// SIDURL SID Request 요청 URL
const SIDURL = "http://%s:%d/angora/v2/query/jobs/%s"

// SendSID SID 전송
func (sidSender *SIDSender) SendSID(sid string) error {
	client := http.Client{}
	var req *http.Request
	var err error

	// Set Header
	token, err := sidSender.Auth.GetAuthToken()
	if err != nil {
		sidSender.log.Errorf("Fail Get Auth Token In Send DSL[ %s ]", err.Error())
		stat.SimStat{}.SendSIDErr()
		return err
	}
	url := fmt.Sprintf(SIDURL, sidSender.Conf.Server.IP, sidSender.Conf.Server.Port, sid)
	req, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		stat.SimStat{}.SendSIDErr()
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Angora "+token)

	stat.SimStat{}.SendSID()
	sidSender.RunningSID[sid] = 1
	resp, err := client.Do(req)
	if err != nil {
		sidSender.log.Debugf("[ SIM >> SERVER ] Error SID Request")
		stat.SimStat{}.SendSIDErr()
		delete(sidSender.RunningSID, sid)
		return nil
	}
	sidSender.ReadSidResponse(resp, sid)
	return nil
}

// ReadSidResponse read sid response
func (sidSender *SIDSender) ReadSidResponse(res *http.Response, sid string) error {
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusOK:
		//result := models.ResponseBody{}
		//decoder := json.NewDecoder(res.Body)
		//err := decoder.Decode(&result)
		//if err != nil && err != io.EOF {
		//	stat.SimStat{}.ErrSID()
		//	delete(sidSender.RunningSID, sid)
		//	sidSender.log.Errorf("[ SIM << SERVER ] Error SID Response")
		//	return err
		//}
		sidSender.log.Debugf("[ SIM << SERVER ] Receive Data?")
		stat.SimStat{}.RecvData()
		delete(sidSender.RunningSID, sid)
		return nil
	default:
		sidSender.log.Debugf("[ SIM << SERVER ] Receive Error[ %d ]", res.StatusCode)
		stat.SimStat{}.ErrSID()
		delete(sidSender.RunningSID, sid)
		return fmt.Errorf("fail SID request with status code[ %d ]", res.StatusCode)
	}
}

// SIDCloseURL SID close Request 요청 URL
const SIDCloseURL = "http://%s:%d/angora/v2/query/jobs/%s/close"

// SendClose send sid close
func (sidSender *SIDSender) SendClose() {
	var token string
	var err error

	// Get Header
	token, err = sidSender.Auth.GetAuthToken()
	if err != nil {
		sidSender.log.Errorf("Fail Get Auth Token In Send DSL[ %s ]", err.Error())
		return
	}

	client := http.Client{}
	var req *http.Request
	var res *http.Response

	for k := range sidSender.RunningSID {

		url := fmt.Sprintf(SIDURL, sidSender.Conf.Server.IP, sidSender.Conf.Server.Port, k)
		req, err = http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			sidSender.log.Errorf("Failed To Create Http.Request[ %s ]", err.Error())
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Angora "+token)

		res, err = client.Do(req)
		if err != nil {
			sidSender.log.Errorf("Failed To Request JOB CLOSE[ %s ]", err.Error())
			return
		}
		res.Body.Close()
	}
}
