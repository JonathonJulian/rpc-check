package main

import (
	"fmt"
	"time"

	"github.com/firstrow/tcp_server"
	"github.com/mackerelio/go-osstat/cpu"
)

func main() {
	server := tcp_server.New(":9876")

	server.OnNewClient(func(c *tcp_server.Client) {
		fmt.Println("Client connected")
		//	cpuIdle, err := getIdleTime()

		// if err != nil {
		// 	fmt.Println(err)
		// 	c.Close()
		// 	return
		// }

		// if cpuIdle < 10 {
		// 	// Set server weight to half
		// 	c.Send("50%\n")
		// } else {
		// 	c.Send("100%\n")
		// }
		c.Send("50%\n")
		c.Close()
	})

	server.Listen()
}

func getIdleTime() (float64, error) {
	before, err := cpu.Get()
	if err != nil {
		return 0, err
	}
	time.Sleep(time.Duration(1) * time.Second)
	after, err := cpu.Get()
	if err != nil {
		return 0, err
	}
	total := float64(after.Total - before.Total)
	cpuIdle := float64(after.Idle-before.Idle) / total * 100
	return cpuIdle, nil
}
