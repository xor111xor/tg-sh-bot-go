package internal

import (
	"fmt"
	"os/exec"
)

type Task struct {
	Pid     int
	CmdText string
	Cmd     *exec.Cmd
}

var Tasks []Task

func (t Task) String() string {
	return fmt.Sprintf("%v, %v", t.Pid, t.CmdText)
}
