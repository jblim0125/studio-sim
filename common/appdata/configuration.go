package appdata

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
)

// Configuration 프로그램 설정 정보
type Configuration struct {
	Log      LogConfiguration    `yaml:"log" json:"log"`
	Server   ServerConfiguration `yaml:"server" json:"server"`
	SendRule DSLSendRule         `yaml:"sendRule" json:"sendRule"`
}

// LogConfiguration  로그 설정 정보
type LogConfiguration struct {
	Output        string `yaml:"output" json:"output"`
	Level         string `yaml:"level" json:"level"`
	SavePath      string `yaml:"savePath" json:"savePath"` // 파일 출력 시 옵션
	SizePerFileMb int32  `yaml:"sizePerFileMb" json:"sizePerFileMb"`
	MaxOfDay      int32  `yaml:"maxOfDay" json:"maxOfDay"`
	MaxAge        int32  `yaml:"maxAge" json:"maxAge"`
	Compress      bool   `yaml:"compress" json:"compress"`
}

// List of supported log output
const (
	LogOutStdout string = "stdout"
	LogOutFile   string = "file"
)

// CheckLogLevel check loglevel and return logrus log level
func CheckLogLevel(lv string) (int, error) {
	switch lv {
	case LvDebug:
		return int(logrus.DebugLevel), nil
	case LvInfo:
		return int(logrus.InfoLevel), nil
	case LvWarn:
		return int(logrus.WarnLevel), nil
	case LvError:
		return int(logrus.ErrorLevel), nil
	case LvSilent:
		return int(logrus.FatalLevel), nil
	default:
		return -1, fmt.Errorf("ERROR. Not Supported Log Level")
	}
}

// List of supported log level
const (
	LvDebug  string = "debug"
	LvInfo   string = "info"
	LvWarn   string = "warn"
	LvError  string = "error"
	LvSilent string = "silent"
)

// ServerConfiguration Server Configuration
type ServerConfiguration struct {
	IP   string `yaml:"ip" json:"ip"`
	Port int    `yaml:"port" json:"port"`
}

// DSLSendRule DSL 전송 룰
type DSLSendRule struct {
	Encrypt    bool `yaml:"encrypt" json:"encrypt"`
	NumThread  int  `yaml:"numThread" json:"numThread"`
	Period     int  `yaml:"period" json:"period"`
	NumSend    int  `yaml:"numSend" json:"numSend"`
	PeriodDSL  int  `yaml:"periodDSL" json:"periodDSL"`
	Infinite   bool `yaml:"infinite" json:"infinite"`
	RunningDSL int  `yaml:"runningDSL" json:"runningDSL"`
}

// Print 환경 설정 주어진 io.writer로 출력
func (c *Configuration) Print(writer io.Writer) {
	enc := json.NewEncoder(writer)
	enc.SetIndent("", "    ")
	if err := enc.Encode(c); err != nil {
		writer.Write([]byte(err.Error()))
		return
	}
}
