package main

import (
	"os"

	"github.com/bot-1337/replay/pkg/replay"
	"github.com/sirupsen/logrus"
)

func main() {
	err := replay.App.Run(os.Args)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
