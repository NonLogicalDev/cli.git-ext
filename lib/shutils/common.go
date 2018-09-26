package shutils

import (
	"os/exec"
	"fmt"
)

type ShCMD struct {
	X *exec.Cmd
}

func Cmd(name string, args ...interface{}) *ShCMD {
	var strArgs []string
	for _, arg := range args {
		strArgs = append(strArgs, fmt.Sprintf("%v", arg))
	}
	return &ShCMD{exec.Command(name, strArgs...)}
}

