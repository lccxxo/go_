package alarm

import "fmt"

// EmailNotifier
type EmailNotifier struct {
	Recipient string
}

func (e *EmailNotifier) Notify(message string) {
	fmt.Printf("发送电子邮件到 %s: %s\n", e.Recipient, message)
}
