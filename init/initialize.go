package init

import (
	"github.com/jblim0125/cache-sim/internal"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/jblim0125/cache-sim/common"
	"github.com/jblim0125/cache-sim/common/appdata"
	"github.com/jblim0125/cache-sim/internal/dsl"
	"github.com/jblim0125/cache-sim/internal/stat"
	"github.com/jblim0125/cache-sim/internal/worker"
)

// Context main context
type Context struct {
	Log  *common.Logger
	Conf *appdata.Configuration
	DSLs *[]string
	//Auth   *internal.Auth
	Sender []*worker.Sender
	Auth   *internal.Auth
}

// New create context instance
func (Context) New() *Context {
	return new(Context)
}

func (c *Context) InitLog() {
	log := common.Logger{}.GetInstance()
	log.Start()
	c.Log = log
}

func (c *Context) ReadConfig(path string) error {
	log := common.Logger{}.GetInstance()
	configManager := common.ConfigManager{}.New(log)
	// Path
	configPath := filepath.Dir(path)
	configName := strings.TrimLeft(path, configPath)
	configType := "yaml"
	// Config file struct
	conf := new(appdata.Configuration)
	// Read
	if err := configManager.ReadConfig(configPath, configName, configType, conf); err != nil {
		return err
	}
	c.Conf = conf
	log.Error("Running Option")
	conf.Print(c.Log.Out)
	log.Errorf("[ Configuration ] Read ............................................................ [ OK ]")
	return nil
}

// SetLogLevel set log level
func (c *Context) SetLogLevel(input string) {
	log := common.Logger{}.GetInstance()
	log.SetLogLevelByString(input)
}

func (c *Context) ReadDSLs(dslPath string) error {
	//dsls, err := dsl.ReadSampleDSL(dslPath)
	//if err != nil {
	//	return err
	//}
	dsls, err := dsl.ReadPlainDSL(dslPath)
	if err != nil {
		return err
	}
	c.Log.Errorf("Read DSL Cont[ %d ]", len(dsls))
	c.DSLs = &dsls
	return nil
}

// Initialize simulator application initialize
func (c *Context) Initialize(dslPath string) error {

	auth, err := internal.Auth{}.Initialize(c.Conf.Server.IP, c.Conf.Server.Port)
	if err != nil {
		return err
	}
	c.Auth = auth

	for i := 0; i < c.Conf.SendRule.NumThread; i++ {
		sender := worker.Sender{}.NewSender(c.Log, c.Conf, c.Auth, dslPath)
		c.Sender = append(c.Sender, sender)
	}
	c.Log.Errorf("[ ALL ] Initialize ................................................................ [ OK ]")
	return nil
}

// StartSubModules Start SubModule And Waiting Signal / Main Loop
func (c *Context) StartSubModules() {
	// Signal
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM)
	c.Log.Errorf("[ Signal ] Listener Start ......................................................... [ OK ]")

	//c.Log.Infof("Total DSL Count [ %d ]", len(*c.DSLs))

	// Start Sender
	for id := 0; id < c.Conf.SendRule.NumThread; id++ {
		go c.Sender[id].Run(id)
	}
	//c.Log.Errorf("[ Router ] Listener Start ......................................................... [ OK ]")

	for {
		select {
		case sig := <-signalChannel:
			c.Log.Errorf("[ SIGNAL ] Receive [ %s ]", sig.String())
			c.StopSubModules()
			return
		case <-time.After(time.Second * 1):
			// 메인 Goroutine에서 주기적으로 무언가를 동작하게 하고 싶다면? 다음을 이용
			// 예 : 성능 시험과 같이 로그가 아닌 통계적인 부분만으로 상태를 체크해야 한다면?
			// c.Log.Errorf("Running...")
			stat.SimStat{}.Print(c.Log.Logger)
		}
	}
}

// StopSubModules Stop Submodules
func (c *Context) StopSubModules() {
	for id := 0; id < c.Conf.SendRule.NumThread; id++ {
		c.Sender[id].Destroy()
	}
	for id := 0; id < c.Conf.SendRule.NumThread; id++ {
		c.Sender[id].DestroyWait()
	}
	c.Log.Errorf("[ Sub goRoutine ] Shutdown ........................................................ [ OK ]")
}
