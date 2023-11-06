package internal

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	tb "gopkg.in/telebot.v3"
)

// var Bot *tb.Bot

func HandleHelp(ctx tb.Context) error {
	message := `
  Any input text will call as shell commend.
  Support command:
  /tasks show all running tasks`
	return ctx.Send(message)
}

func HandleTasks(ctx tb.Context) error {
	tasksText := []string{}

	for i := range Tasks {
		tasksText = append(tasksText, Tasks[i].String())
	}
	msg := strings.Join(tasksText, "\r\n")
	if len(msg) == 0 {
		msg = "Task list is empty"
	}
	return ctx.Send(msg)
}

func doCd(ctx tb.Context) bool {
	cmd := ctx.Text()
	if !strings.HasPrefix(strings.ToLower(cmd), "cd ") {
		return false
	}
	err := os.Chdir(cmd[3:])
	if err != nil {
		_ = ctx.Send(err)
		return false
	}
	pwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	msg := fmt.Sprintf("pwd: %s", pwd)
	_ = ctx.Send(msg)
	if err != nil {
		log.Println(err)
	}
	return true
}

func HandleExecCommand(ctx tb.Context) error {
	if doCd(ctx) {
		return nil
	}
	commandText := ctx.Text()
	doExecCommand(commandText, ctx)
	return nil
}

func replyCmdOut(out string, ctx tb.Context) error {
	if len(out) == 0 {
		return nil
	}
	if len(out) > 500 {
		out = out[:500]
	}
	return ctx.Send(out)
}

func doExecCommand(commandText string, ctx tb.Context) {
	cmd := exec.Command("/bin/sh", "-c", commandText)
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		log.Println(err)
		return
	}

	err = cmd.Start()
	if err != nil {
		log.Println(err)
		return
	}

	task := Task{cmd.Process.Pid, commandText, cmd}
	Tasks = append(Tasks, task)

	reader := bufio.NewReader(stdout)
	out := ""
	idx := 0

	startTime := time.Now()
	leapTime := time.Duration(1) * time.Second

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		out += line
		if time.Since(startTime) > leapTime {
			err := replyCmdOut(out, ctx)
			if err != nil {
				log.Println(err)
			}
			idx += 1
			out = ""
			startTime = time.Now()
		}
		if idx > 3 {
			msg := fmt.Sprintf("Command not finished, you can kill by send kill %d", task.Pid)
			_ = ctx.Send(msg)
		}
	}
	err = cmd.Wait()
	if err != nil {
		log.Println(err)
	}

	// Tasks.remove(task)
	for i := 0; i < len(Tasks); i++ {
		if Tasks[i] == task {
			Tasks = append(Tasks[:i], Tasks[i+1:]...)
			i--
		}
	}
	err = replyCmdOut(out, ctx)
	if err != nil {
		log.Println(err)
	}

	if idx > 3 {
		msg := fmt.Sprintf("Task finished: %s", task.CmdText)
		_ = ctx.Send(msg)
	}
}

func RunHandlers(settings tb.Settings) error {
	bot, err := tb.NewBot(settings)
	if err != nil {
		log.Fatal(err)
		return err
	}

	bot.Handle("/help", HandleHelp)
	bot.Handle("/start", HandleHelp)
	bot.Handle("/tasks", HandleTasks)
	bot.Handle(tb.OnText, HandleExecCommand)

	bot.Start()
	return nil
}
