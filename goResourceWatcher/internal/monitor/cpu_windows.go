// cpu_windows.go
//go:build windows

package monitor

import (
	"context"
	"fmt"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/go_/lccxxo/goResourceWatcher/internal/alarm"
	"github.com/go_/lccxxo/goResourceWatcher/internal/logger"
)

type CPUUsage struct {
	Alarm *alarm.Alarm
}

func NewCPUUsage(alarm *alarm.Alarm) *CPUUsage {
	return &CPUUsage{Alarm: alarm}
}

func (c *CPUUsage) monitor(ctx context.Context, pause *sync.Cond, isPaused *bool) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("任务结束")
			return
		default:
			pause.L.Lock()
			for *isPaused {
				pause.Wait() // 等待被唤醒
			}
			pause.L.Unlock()

			// 执行任务逻辑
			usage, err := c.getCPUUsage()
			if err != nil {
				fmt.Println("get cpu usage error:", err)
				return
			}
			//fmt.Printf("cpu usage: %.2f%%\n", usage)
			if usage > 30.0 {
				c.Alarm.NotifyObservers(fmt.Sprintf("cpu usage is too high: %.2f%%", usage))
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (c *CPUUsage) getCPUUsage() (float64, error) {
	idle1, kernel1, user1, err := c.getSystemTimes()
	if err != nil {
		logger.Logger.Errorf("get windows system time error:%s", err)
		return 0, err
	}

	time.Sleep(1000 * time.Millisecond)

	idle2, kernel2, user2, err := c.getSystemTimes()
	if err != nil {
		logger.Logger.Errorf("get windows system time error:%s", err)
		return 0, err
	}

	totalIdle := idle2 - idle1
	totalKernel := kernel2 - kernel1
	totalUser := user2 - user1

	total := totalKernel + totalUser
	used := total - totalIdle

	return float64(used) / float64(total) * 100, nil
}

func (c *CPUUsage) getSystemTimes() (idle, kernel, user uint64, err error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")           //kernel32.dll包是windows操作系统的一个库
	procGetSystemTimes := kernel32.NewProc("GetSystemTimes") // GetSystemTimes获取cpu时间

	var idleTime, kernelTime, userTime syscall.Filetime

	ret, _, err := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&idleTime)),
		uintptr(unsafe.Pointer(&kernelTime)),
		uintptr(unsafe.Pointer(&userTime)),
	)
	if ret == 0 {
		logger.Logger.Infof("don't get windows system times")
		return 0, 0, 0, err
	}

	idle = c.fileTimeToUint64(&idleTime)
	kernel = c.fileTimeToUint64(&kernelTime)
	user = c.fileTimeToUint64(&userTime)

	return idle, kernel, user, nil
}

func (c *CPUUsage) fileTimeToUint64(ft *syscall.Filetime) uint64 {
	return uint64(ft.HighDateTime)<<32 + uint64(ft.LowDateTime)
}
