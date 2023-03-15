package fuzzer

import (
	"bytes"
	"os"
	"os/exec"
)

type Executor struct {
}

type Output struct {
	Err   error
	O     string
	Trace string
}

type Input struct {
	c    *Chain
	cmd  string
	args []string
}

func (e *Executor) Run(in Input) Output {
	command := exec.Command(in.cmd, in.args...)
	var instr string
	if in.c == nil {
		instr = "Input="
	} else {
		instr = "Input=" + in.c.ToString()
	}
	command.Env = append(os.Environ(), instr)
	//给标准输入以及标准错误初始化一个buffer，每条命令的输出位置可能是不一样的，
	//比如有的命令会将输出放到stdout，有的放到stderr
	command.Stdout = &bytes.Buffer{}
	command.Stderr = &bytes.Buffer{}
	//执行命令，直到命令结束
	err := command.Run()
	//打印命令行的标准输出
	return Output{
		err,
		command.Stdout.(*bytes.Buffer).String(),
		command.Stderr.(*bytes.Buffer).String(),
	}
}
