package main

import (
	"fmt"
	"github.com/jblim0125/studio-sim/internal"
	"github.com/jblim0125/studio-sim/internal/dsl"
	"github.com/jblim0125/studio-sim/internal/stat"
	"github.com/jblim0125/studio-sim/internal/worker"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jblim0125/studio-sim/common"
	"github.com/jblim0125/studio-sim/common/appdata"
	"github.com/sirupsen/logrus"
)

// Context main context
type Context struct {
	Log       *common.Logger
	CM        *common.ConfigManager
	Conf      *appdata.Configuration
	DSLs      *map[string]interface{}
	Auth      *internal.Auth
	DSLSender *worker.DSLSender
	SIDSender *worker.SIDSender
}

// InitLog Initialize logger
func (c *Context) InitLog() {
	log := common.Logger{}.GetInstance()
	log.SetLogLevel(logrus.DebugLevel)
	c.Log = log
}

// ReadConfig Read Configuration File By Viper
func (c *Context) ReadConfig() error {
	c.CM = common.ConfigManager{}.New(c.Log.Logger)
	// Write Config File Info
	configPath := "./configs"
	configName := "dev"
	configType := "yaml"
	// Config file struct
	conf := new(appdata.Configuration)
	// Read
	if err := c.CM.ReadConfig(configPath, configName, configType, conf); err != nil {
		return err
	}
	// Save
	c.Conf = conf

	// Set Watcher
	c.CM.SetOnChanged(configPath, configName, configType,
		func(conf interface{}) {
			newConf := conf.(*appdata.Configuration)
			c.Log.Infof("%+v\n", newConf)
		}, conf)
	//c.Log.Debugf("%+v", c.Conf)
	c.Log.Errorf("Running Option")
	c.Conf.Print(c.Log.Out)
	c.Log.Errorf("[ Configuration ] Read ............................................................ [ OK ]")
	return nil
}

// SetLogger set log level, log output. and etc
func (c *Context) SetLogger() error {
	if err := c.Log.Setting(&c.Conf.Log); err != nil {
		return err
	}
	return nil
}

// Initialize env/config load and sub moduel init
func Initialize() (*Context, error) {
	c := new(Context)
	c.Conf = new(appdata.Configuration)

	// 환경 변수, 컨피그를 읽어 들이는 과정에서 로그 출력을 위해
	// 아주 기초적인 부분만을 초기화 한다.
	c.InitLog()
	c.Log.Start()

	// Read Config
	if err := c.ReadConfig(); err != nil {
		return nil, err
	}

	// Setting Log(from env and conf)
	if err := c.SetLogger(); err != nil {
		return nil, err
	}

	dsls, err := dsl.ReadSampleDSL("./sample/dsl.json")
	if err != nil {
		return nil, err
	}
	c.DSLs = dsls

	//auth, err := Auth{}.Initialize(c.Conf.Server.IP, c.Conf.Server.Port)
	//if err != nil {
	//	return nil, err
	//}
	//c.Auth = auth

	dslSender := worker.DSLSender{}.NewDSLSender(c.Log, c.Auth, c.DSLs, c.Conf)
	c.DSLSender = dslSender

	sidSender := worker.SIDSender{}.NewSIDSender(c.Log, c.Auth, c.Conf)
	c.SIDSender = sidSender

	c.Log.Errorf("[ ALL ] Initialize ................................................................ [ OK ]")
	return c, nil
}

// StartSubModules Start SubModule And Waiting Signal / Main Loop
func (c *Context) StartSubModules() {
	// Signal
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM)
	c.Log.Errorf("[ Signal ] Listener Start ......................................................... [ OK ]")

	// TODO : Start Other Sub Modules
	for id := 0; id < c.Conf.SendRule.NumThread; id++ {
		go c.DSLSender.Run(id, c.SIDSender.Dsl[id])
		go c.SIDSender.RunSIDReceiver(id)
		go c.SIDSender.RunSIDSender(id)
	}
	//c.Log.Errorf("[ Router ] Listener Start ......................................................... [ OK ]")

	for {
		select {
		//case err := <-echoServerErr:
		//    c.Log.Errorf("[ SERVER ] ERROR[ %s ]", err.Error())
		//    c.StopSubModules()
		//    return
		case sig := <-signalChannel:
			c.Log.Errorf("[ SIGNAL ] Receive [ %s ]", sig.String())
			c.StopSubModules()
			return
		case <-time.After(time.Second * 5):
			// 메인 Goroutine에서 주기적으로 무언가를 동작하게 하고 싶다면? 다음을 이용
			// 예 : 성능 시험과 같이 로그가 아닌 통계적인 부분만으로 상태를 체크해야 한다면?
			// c.Log.Errorf("Running...")
			stat.SimStat{}.Print(c.Log.Logger)
		}
	}
}

// StopSubModules Stop Submodules
func (c *Context) StopSubModules() {
	//integration.JhmsClient{}.Destroy()
	//c.Log.Errorf("[ JHMS Client ] Shutdown .......................................................... [ OK ]")
}

// @title Cache Server API
// @version 1.0.0
// @description This is a cache server.

// @contact.name API Support
// @contact.url http://mobigen.com
// @contact.email irisdev@mobigen.com

// @host localhost:8080
// @BashPath /
func main() {
	// Initialize Sub module And Read Env, Config
	c, err := Initialize()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Start sub module and main loop
	c.StartSubModules()

	// Bye Bye
	c.Log.Shutdown()
}
