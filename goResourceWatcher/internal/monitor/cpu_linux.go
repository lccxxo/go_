// cpu_linux.go
//go:build linux

package monitor

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

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

			if usage > 30.0 {
				c.Alarm.NotifyObservers(fmt.Sprintf("cpu usage is too high: %.2f%%", usage))
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (c *CPUUsage) getCPUUsage() (float64, error) {
	idle1, kernel1, user1, irq1, softirq1, steal1, err := c.getSystemTimes()
	if err != nil {
		logger.Logger.Errorf("get linux system time error:%s", err)
		return 0, err
	}

	time.Sleep(1 * time.Second)

	idle2, kernel2, user2, irq2, softirq2, steal2, err := c.getSystemTimes()
	if err != nil {
		logger.Logger.Errorf("get linux system time error:%s", err)
		return 0, err
	}

	totalIdle := float64(idle2 - idle1)
	totalKernel := float64(kernel2 - kernel1)
	totalUser := float64(user2 - user1)
	totalIrq := float64(irq2 - irq1)
	totalSoftIrq := float64(softirq2 - softirq1)
	totalSteal := float64(steal2 - steal1)

	total := totalKernel + totalUser + totalIrq + totalSoftIrq + totalSteal + totalIdle
	used := total - totalIdle

	if total == 0 {
		return 0, nil
	}

	return (used / total) * 100, nil
}

func (c *CPUUsage) getSystemTimes() (idle, kernel, user, irq, softirq, steal uint64, err error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		logger.Logger.Errorf("read file /proc/stat error:%s", err)
		return 0, 0, 0, 0, 0, 0, err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 1 {
		return 0, 0, 0, 0, 0, 0, err
	}

	fields := strings.Fields(lines[0])
	if len(fields) < 9 {
		return 0, 0, 0, 0, 0, 0, err
	}

	// 用户态时间包括user、nice，分别是用户态CPU时间、低优先级用户态CPU时间（进程nice值为1-19）
	user, err = strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		logger.Logger.Errorf("Error Obtaining cpu user mode usage time:%s", err)
		return 0, 0, 0, 0, 0, 0, err
	}
	nice, err := strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		logger.Logger.Errorf("Error Obtaining cpu nice mode usage time:%s", err)
		return 0, 0, 0, 0, 0, 0, err
	}
	user += nice
	// 内核态执行时间
	kernel, err = strconv.ParseUint(fields[3], 10, 64)
	if err != nil {
		logger.Logger.Errorf("Error Obtaining cpu kernel mode usage time:%s", err)
		return 0, 0, 0, 0, 0, 0, err
	}
	// CPU闲置时间包括idle和iowait，分别是CPU空闲时间、等待I/O操作完成时CPU的空闲时间
	idle, err = strconv.ParseUint(fields[4], 10, 64)
	if err != nil {
		logger.Logger.Errorf("Error Obtaining cpu idle mode usage time:%s", err)
		return 0, 0, 0, 0, 0, 0, err
	}
	iowait, err := strconv.ParseUint(fields[5], 10, 64)
	if err != nil {
		logger.Logger.Errorf("Error Obtaining cpu iowait mode usage time:%s", err)
		return 0, 0, 0, 0, 0, 0, err
	}
	idle += iowait
	// 硬中断CPU时间
	irq, err = strconv.ParseUint(fields[6], 10, 64)
	if err != nil {
		logger.Logger.Errorf("Error Obtaining cpu irq mode usage time:%s", err)
		return 0, 0, 0, 0, 0, 0, err
	}
	// 软中断CPU时间
	softirq, err = strconv.ParseUint(fields[7], 10, 64)
	if err != nil {
		logger.Logger.Errorf("Error Obtaining cpu softirq mode usage time:%s", err)
		return 0, 0, 0, 0, 0, 0, err
	}
	// 被系统监控程序使用的CPU时间
	steal, err = strconv.ParseUint(fields[8], 10, 64)
	if err != nil {
		logger.Logger.Errorf("Error Obtaining cpu steal mode usage time:%s", err)
		return 0, 0, 0, 0, 0, 0, err
	}

	return idle, kernel, user, irq, softirq, steal, err
}
