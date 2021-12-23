package worker

import (
	"fmt"
	"github.com/jblim0125/studio-sim/common"
	"github.com/jblim0125/studio-sim/common/appdata"
	"github.com/jblim0125/studio-sim/internal"
	"github.com/jblim0125/studio-sim/internal/stat"
	"github.com/jblim0125/studio-sim/models"
	"net/http"
)

// SIDSender SID 전송 객체
type SIDSender struct {
	log  *common.Logger
	Dsl  []chan models.HTTPData
	Sid  []chan string
	Auth *internal.Auth
	Conf *appdata.Configuration
}

// NewSIDSender SID 전송 goroutine 생성
func (SIDSender) NewSIDSender(log *common.Logger,
	auth *internal.Auth, conf *appdata.Configuration) *SIDSender {
	sidSender := &SIDSender{
		log:  log,
		Auth: auth,
		Conf: conf,
	}

	for i := 0; i < conf.SendRule.NumThread; i++ {
		dslCh := make(chan models.HTTPData)
		sidSender.Dsl = append(sidSender.Dsl, dslCh)
		sidCh := make(chan string)
		sidSender.Sid = append(sidSender.Sid, sidCh)
	}
	return sidSender
}

// RunSIDReceiver angora/query/job 의 응답을 수신하고, 결과를 SID 전송 goroutine으로 전달
func (sidSender *SIDSender) RunSIDReceiver(id int) {
	sidSender.log.Errorf("[ DSL Response Receiver [ %d ] Start ............................................... [ OK ]", id)
	for {
		select {
		case resp := <-sidSender.Dsl[id]:
			err := sidSender.ReadDSLResponse(id, resp)
			if err != nil {
				sidSender.log.Errorf("SID Receive Err[ %s ]", err.Error())
			}
		}
	}
}

// ReadDSLResponse DSL 응답 메시지 확인
func (sidSender *SIDSender) ReadDSLResponse(id int, resp models.HTTPData) error {
	defer resp.Response.Body.Close()
	switch resp.Response.StatusCode {
	case http.StatusOK:
		//res := models.DSLResponseBody{}
		//decoder := json.NewDecoder(resp.Response.Body)
		//err := decoder.Decode(&res)
		//if err != nil && err != io.EOF {
		//	stat.SimStat{}.ErrDSL()
		//	return err
		//}
		//sidSender.log.Debugf("URL : %s", resp.Response.Request.URL.String())
		stat.SimStat{}.RecvSID()
		//sidSender.Sid[id] <- res.SID
		sidSender.Sid[id] <- "test"
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
		case sid := <-sidSender.Sid[id]:
			go func() {
				err := sidSender.SendSID(sid)
				if err != nil {
					sidSender.log.Errorf("Send SID Err[ %s ]", err.Error())
				}
			}()
		}
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
		stat.SimStat{}.SendDSLErr()
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Angora "+token)

	url := fmt.Sprintf(SIDURL, sidSender.Conf.Server.IP, sidSender.Conf.Server.Port, sid)

	req, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		stat.SimStat{}.SendSIDErr()
		return nil
	}
	resp, err := client.Do(req)
	if err != nil {
		stat.SimStat{}.SendSIDErr()
		return nil
	}
	stat.SimStat{}.SendSID()

	defer resp.Body.Close()

	// TODO : Check Response Code? Body?
	return nil
}
