package utils

import (
	"os"

	logrus "github.com/Sirupsen/logrus"
)

// StopOnErr exits on any error after logging it.
func StopOnErr(err error) {
	if err == nil {
		return
	}
	defer os.Exit(-1)

	newMessage := err.Error()
	if newMessage != "" {
		logrus.Error(newMessage)
	}
}
