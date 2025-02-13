package main

import (
	"fmt"
	"github.com/go_/lccxxo/goResourceWatcher/internal/startup"
)

const watcher = `
   __
  /  \
 |    |
  \__/
Watcher is active
`

func main() {
	fmt.Print(watcher)
	server := startup.Server{}
	server.StartUp()
	server.HandleSignal()

	select {}
}
