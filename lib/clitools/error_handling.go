package clitools

import (
	"fmt"
	"os"
)

func UserErrorStr(t, msg string, args ...interface{}) {
	fmt.Printf("Error [%v]: %v\n", t, fmt.Sprintf(msg, args...))
	os.Exit(128)
}

func UserError(err error) {
	if err != nil {
		UserErrorStr("Go", err.Error())
	}
}

func UserErrorWrap(err error, format string, args ...interface{}) {
	if err != nil {
		UserErrorStr("Go", fmt.Sprintf("%v |%v:", err.Error(), fmt.Sprintf(format, args...)))
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
