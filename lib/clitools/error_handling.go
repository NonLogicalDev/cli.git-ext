package clitools

import (
	"fmt"
	"os"
)

func UserErrorStr(t, msg string) {
	fmt.Printf("Error [%v]: %v\n", t, msg)
	os.Exit(128)
}

func UserError(err error) {
	if err != nil {
		UserErrorStr("Go", err.Error())
	}
}

func UserFriendlyPanic(allowPanic bool) {
	if allowPanic {
		return
	}
	if r := recover(); r != nil {
		if err, ok := r.(error); ok {
			UserError(err)
		}
	}
}

