package init

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jblim0125/cache-sim/common"
	"github.com/jblim0125/cache-sim/common/appdata"
	"github.com/jblim0125/cache-sim/internal"
	"github.com/jblim0125/cache-sim/internal/dsl"
	"github.com/jblim0125/cache-sim/internal/stat"
	"github.com/jblim0125/cache-sim/internal/worker"
)

// Context main context
type Context struct {
	Log    *common.Logger
	Conf   *appdata.Configuration
	DSLs   *map[string]interface{}
	Auth   *internal.Auth
	Sender *worker.Sender
}

// New create context instance
func (Context) New(log *common.Logger, config *appdata.Configuration) *Context {
	if log == nil || config == nil {
		return nil
	}
	return &Context{
		Log:  log,
		Conf: config,
	}
}

// Initialize simulator application initialize
func (c *Context) Initialize(dslPath string) error {
	dsls, err := dsl.ReadSampleDSL(dslPath)
	if err != nil {
		return err
	}
	c.Log.Errorf("Read DSL Cont[ %d ]", len(*dsls))
	c.DSLs = dsls

	auth, err := internal.Auth{}.Initialize(c.Conf.Server.IP, c.Conf.Server.Port)
	if err != nil {
		return err
	}
	c.Auth = auth

	sender := worker.Sender{}.NewSender(c.Log, c.Auth, c.DSLs, c.Conf)
	c.Sender = sender
	c.Log.Errorf("[ ALL ] Initialize ................................................................ [ OK ]")
	return nil
}

// StartSubModules Start SubModule And Waiting Signal / Main Loop
func (c *Context) StartSubModules() {
	// Signal
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM)
	c.Log.Errorf("[ Signal ] Listener Start ......................................................... [ OK ]")

	// TODO : Start Other Sub Modules
	for id := 0; id < c.Conf.SendRule.NumThread; id++ {
		go c.Sender.Run(id)
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
	c.Sender.Destroy()
	c.Log.Errorf("[ Sub goRoutine ] Shutdown ........................................................ [ OK ]")
}
