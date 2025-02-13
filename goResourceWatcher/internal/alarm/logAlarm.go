package alarm

import "github.com/go_/lccxxo/goResourceWatcher/internal/logger"

// LogNotifier
type LogNotifier struct{}

func (l *LogNotifier) Notify(message string) {
	logger.Logger.Errorf("日志告警: %s\n", message)
}
