package main

import (
	"log"
)

func log_err(msg string, err error, panic bool) {
	if !panic {
		log.Printf("%s\n%v", msg, err)
	}
	if panic {
		log.Panicf("%s\n%v", msg, err)
	}
}
