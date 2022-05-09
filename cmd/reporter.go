package main

import (
	log "github.com/sirupsen/logrus"
	"reporter/internal/report"
)

func main() {
	log.Info("Start app....")
	config := &report.Config{}
	err := config.Init()
	if err != nil {
		log.Fatalln(err)
	}
	app := report.NewApp(config)
	go app.Run()
	exit := make(chan int)
	<-exit
}
