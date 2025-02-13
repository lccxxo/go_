package startup

import (
	"fmt"
	"github.com/go_/lccxxo/goResourceWatcher/internal/database"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/go_/lccxxo/goResourceWatcher/config"
	"github.com/go_/lccxxo/goResourceWatcher/internal/alarm"
	"github.com/go_/lccxxo/goResourceWatcher/internal/logger"
	"github.com/go_/lccxxo/goResourceWatcher/internal/monitor"
	"github.com/go_/lccxxo/goResourceWatcher/internal/task"
)

type Server struct {
}

func (s *Server) StartUp() {
	// init config
	config.LoadConfig()
	// init logger
	logger.InitLogger()
	// init mysql
	database.InitDatabase()
	// init task entry
	task.AutoMigrateTaskEntity()
	// init taskQueue
	taskQueue := task.NewTaskQueue()
	// init alarm
	newAlarm := alarm.NewAlarm()
	emailNotifier := alarm.EmailNotifier{Recipient: "majunhong@gmail.com"}
	logNotifier := alarm.LogNotifier{}
	newAlarm.RegisterObservers(&emailNotifier, &logNotifier)
	// init cpu monitor
	cpuMonitor := monitor.NewCPUMonitor(monitor.NewCPUUsage(newAlarm))

	taskQueue.AddTask("CPU usage monitor", cpuMonitor.GetCPUUsage)

	env := os.Getenv("APP_ENV")
	goos := runtime.GOOS
	logger.Logger.Infof("Project is running, the project environment is %s, the project os is %s", env, goos)
}

func (s *Server) Shutdown() {
	err := task.ClearTaskTable()
	if err != nil {
		logger.Logger.Error("Failed to task clear task table:", err)
		return
	}
}

// 关闭系统
func (s *Server) HandleSignal() {
	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGHUP)
	go func() {
		fmt.Println("接受关闭信号")
		<-closeChan
		s.Shutdown()
		os.Exit(0)
	}()
}
