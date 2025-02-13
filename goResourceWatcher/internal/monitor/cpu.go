package monitor

import (
	"context"
	"sync"
)

type CPUMonitorStrategy interface {
	monitor(ctx context.Context, pause *sync.Cond, isPaused *bool)
}

type CPUMonitor struct {
	Strategy CPUMonitorStrategy
	cpuUsage CPUUsage
}

func NewCPUMonitor(usage *CPUUsage) *CPUMonitor {
	return &CPUMonitor{Strategy: usage}
}

func (c *CPUMonitor) GetCPUUsage(ctx context.Context, pause *sync.Cond, isPaused *bool) {
	c.Strategy.monitor(ctx, pause, isPaused)
	return
}
