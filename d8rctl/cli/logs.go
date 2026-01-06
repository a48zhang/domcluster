package cli

import (
	"bytes"
	"d8rctl/config"
	"fmt"
	"io"
	"os"
)

const bufferSize = 4096

// getLogFile 获取日志文件路径
func getLogFile() string {
	return config.GetLogFile()
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

	if lines > 0 {
		logLines, err := readLastLines(file, lines)
		if err != nil {
			return fmt.Errorf("failed to read log file: %w", err)
		}

		for _, line := range logLines {
			fmt.Println(line)
		}
	} else {
		if _, err := io.Copy(os.Stdout, file); err != nil {
			return fmt.Errorf("failed to read log file: %w", err)
		}
	}

	return nil
}

// readLastLines 从文件末尾向前读取指定行数
func readLastLines(file *os.File, lines int) ([]string, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := stat.Size()
	if fileSize == 0 {
		return []string{}, nil
	}

	// 如果文件很小，直接读取全部内容
	if fileSize <= int64(bufferSize) {
		content, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		return extractLastLines(string(content), lines), nil
	}

	// 大文件：从末尾向前读取
	var buf []byte
	var lineCount int
	var offset int64 = fileSize

	for offset > 0 {
		chunkSize := bufferSize
		if offset < int64(chunkSize) {
			chunkSize = int(offset)
		}
		offset -= int64(chunkSize)

		_, err := file.Seek(offset, io.SeekStart)
		if err != nil {
			return nil, err
		}

		chunk := make([]byte, chunkSize)
		_, err = io.ReadFull(file, chunk)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return nil, err
		}

		// 从后向前统计换行符
		for i := chunkSize - 1; i >= 0; i-- {
			if chunk[i] == '\n' {
				lineCount++
				if lineCount > lines {
					// 找到足够的换行符，保留从当前位置开始的所有内容
					buf = append(chunk[i+1:], buf...)
					return extractLastLines(string(buf), lines), nil
				}
			}
		}

		// 将当前chunk添加到缓冲区前面
		buf = append(chunk, buf...)
	}

	// 已经到达文件开头，返回所有内容
	return extractLastLines(string(buf), lines), nil
}

// extractLastLines 从字符串中提取最后几行
func extractLastLines(content string, lines int) []string {
	if content == "" {
		return []string{}
	}

	lineList := bytes.Split([]byte(content), []byte{'\n'})
	total := len(lineList)

	// 如果最后一行是空的（文件以换行符结尾），去掉它
	if total > 0 && len(lineList[total-1]) == 0 {
		total--
	}

	start := 0
	if total > lines {
		start = total - lines
	}

	result := make([]string, 0, total-start)
	for i := start; i < total; i++ {
		result = append(result, string(lineList[i]))
	}

	return result
}