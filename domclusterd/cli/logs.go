package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

const bufferSize = 4096

// getLogFile 获取日志文件路径 - 强制使用 /var/log/domclusterd/
func getLogFile() string {
	return "/var/log/domclusterd/domclusterd.log"
}

// Logs 查看日志
func Logs(lines int) error {
	logFile := getLogFile()

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
		logLines, err := readLastLines(file, lines)
		if err != nil {
			return fmt.Errorf("failed to read log file: %w", err)
		}

		for _, line := range logLines {
			fmt.Println(line)
		}
	} else {
		// 输出所有内容
		if _, err := io.Copy(os.Stdout, file); err != nil {
			return fmt.Errorf("failed to read log file: %w", err)
		}
	}

	return nil
}

// readLastLines 从文件末尾向前读取指定行数
func readLastLines(file *os.File, n int) ([]string, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := stat.Size()
	if fileSize == 0 {
		return []string{}, nil
	}

	// 从文件末尾开始读取
	var buf []byte
	var lines []string
	var offset int64 = fileSize
	var lineCount int = 0

	// 每次读取 bufferSize 大小的数据
	for offset > 0 && lineCount < n {
		readSize := bufferSize
		if offset < bufferSize {
			readSize = int(offset)
		}
		offset -= int64(readSize)

		_, err := file.Seek(offset, io.SeekStart)
		if err != nil {
			return nil, err
		}

		chunk := make([]byte, readSize)
		_, err = io.ReadFull(file, chunk)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return nil, err
		}

		// 将新读取的数据放在前面
		buf = append(chunk, buf...)

		// 统计换行符数量
		for i := len(buf) - 1; i >= 0; i-- {
			if buf[i] == '\n' {
				lineCount++
				if lineCount > n {
					// 找到足够多的行后，截取剩余部分
					buf = buf[i+1:]
					break
				}
			}
		}
	}

	// 按行分割
	scanner := bufio.NewScanner(bytes.NewReader(buf))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// 返回最后 n 行
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	return lines, nil
}