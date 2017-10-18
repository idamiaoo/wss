package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
}
