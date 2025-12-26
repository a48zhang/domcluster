package shell

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"time"
)

// Result 执行结果
type Result struct {
	Success  bool
	ExitCode int
	Stdout   string
	Stderr   string
	Error    string
}

// Execute 同步执行命令
func Execute(command string, args []string, timeout int) (*Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	result := &Result{
		Success:  err == nil,
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result, nil
}

// OutputLine 输出行
type OutputLine struct {
	Type     string // "stdout" or "stderr"
	Text     string
	Finished bool
	ExitCode int
	Error    string
}

// StreamExecute 流式执行命令
func StreamExecute(command string, args []string, outputChan chan<- OutputLine, timeout int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// 读取 stdout
	go readPipe(stdoutPipe, outputChan, "stdout")
	// 读取 stderr
	go readPipe(stderrPipe, outputChan, "stderr")

	// 等待命令完成
	err = cmd.Wait()

	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	// 发送完成信号
	outputChan <- OutputLine{
		Finished: true,
		ExitCode: exitCode,
	}
	close(outputChan)

	return nil
}

func readPipe(pipe io.ReadCloser, outputChan chan<- OutputLine, outputType string) {
	defer pipe.Close()

	buf := make([]byte, 1024)
	for {
		n, err := pipe.Read(buf)
		if n > 0 {
			outputChan <- OutputLine{
				Type: outputType,
				Text: string(buf[:n]),
			}
		}
		if err != nil {
			break
		}
	}
}