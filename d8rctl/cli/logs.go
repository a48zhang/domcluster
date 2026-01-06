package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

const logFile = "d8rctl.log"
const bufferSize = 4096

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

	for offset > 0 && lineCount <= lines {
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
		foundEnoughLines := false
		for i := chunkSize - 1; i >= 0; i-- {
			if chunk[i] == '\n' {
				lineCount++
				if lineCount > lines {
					// 找到足够的换行符，只保留当前 chunk 中从第 lines+1 个换行符之后的部分
					// 此时不需要之前读取的 chunk 数据，因为我们已经找到了足够的换行符
					buf = chunk[i+1:]
					foundEnoughLines = true
					break
				}
			}
		}

		// 如果还没有找到足够的换行符，继续向前读取
		if !foundEnoughLines {
			buf = append(chunk, buf...)
		} else {
			// 已经找到足够的换行符，停止读取更多数据
			break
		}
	}

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