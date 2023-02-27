package fuzzer

import (
	"bytes"
	"os/exec"
)

type Executor struct {
}

type Output struct {
	err error
	o   string
}

func (e *Executor) Run(cmd string, args []string) Output {
	command := exec.Command(cmd, args...)
	//给标准输入以及标准错误初始化一个buffer，每条命令的输出位置可能是不一样的，
	//比如有的命令会将输出放到stdout，有的放到stderr
	command.Stdout = &bytes.Buffer{}
	command.Stderr = &bytes.Buffer{}
	//执行命令，直到命令结束
	err := command.Run()
	if err != nil {
		return Output{
			err,
			command.Stderr.(*bytes.Buffer).String(),
		}
	}
	//打印命令行的标准输出
	return Output{
		nil,
		command.Stdout.(*bytes.Buffer).String(),
	}
}
