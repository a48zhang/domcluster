package cli

import (
	"fmt"
	"io"
	"os"
)

const logFile = "d8rctl.log"

// Logs 查看日志
func Logs(lines int) error {
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return fmt.Errorf("log file not found")
	}

	file, err := os.Open(logFile)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// 读取最后几行
	if lines > 0 {
		// 读取所有内容
		content, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("failed to read log file: %w", err)
		}

		// 按行分割
		logLines := splitLines(string(content))
		total := len(logLines)

		// 获取最后几行
		start := 0
		if total > lines {
			start = total - lines
		}

		for i := start; i < total; i++ {
			fmt.Println(logLines[i])
		}
	} else {
		// 输出所有内容
		if _, err := io.Copy(os.Stdout, file); err != nil {
			return fmt.Errorf("failed to read log file: %w", err)
		}
	}

	return nil
}

// splitLines 按行分割字符串
func splitLines(s string) []string {
	lines := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}