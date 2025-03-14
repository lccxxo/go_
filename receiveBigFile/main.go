package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"sync"
	"time"
)

/*
	作用：接收大文件时对上传大小以及百分比进行记录
*/

// progressReader 跟踪上传的大小
type progressReader struct {
	io.Reader
	total     int64
	read      int64
	mu        sync.Mutex
	StartTime time.Time
}

func (pr *progressReader) Write(p []byte) (n int, err error) {
	pr.mu.Lock()
	pr.read += int64(len(p))
	pr.mu.Unlock()
	return len(p), nil
}

// 获取上传百分比以及已经上传的大小
func (pr *progressReader) getProgress() (float64, int64) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	if pr.total == 0 {
		return 0, 0
	}
	return float64(pr.read) / float64(pr.total) * 100, pr.read
}

func main() {
	filePath := "../uploadBigFile/uploads/pycharm-professional-2024.2.3.exe"
	baseUrl := "http://localhost:8080/upload"

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	
	pr := &progressReader{
		total:     fileSize,
		StartTime: time.Now(),
	}

	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				percent, sent := pr.getProgress()
				log.Printf("进度: %.2f%% 已发送: %s/%s 速率: %s/s",
					percent,
					formatBytes(sent),
					formatBytes(fileSize),
					formatBytes(int64(float64(sent)/time.Since(pr.StartTime).Seconds())),
				)
			case <-done:
				percent, sent := pr.getProgress()
				log.Printf("最终进度: %.2f%% (%s/%s)",
					percent,
					formatBytes(sent),
					formatBytes(pr.total),
				)
				return
			}
		}
	}()

	//	发送请求
	// io.Pipe()方法返回一个read和一个writer read实时读取writer写入的数据
	body, writer := io.Pipe()
	multipartWriter := multipart.NewWriter(writer)
	go func() {
		defer writer.Close()
		defer multipartWriter.Close()

		// 构造文件请求的表单格式
		part, _ := multipartWriter.CreateFormFile("file", fileInfo.Name())
		// 该方法是监控发送进度的重要方法 file为数据源 pr为数据接收器
		// 返回了一个reader 该reader读取的就是file的内容 等待被读取
		// 当io.TeeReader返回的reader被读取时，会触发pr的写入效果，数据不会真正的写入pr，而是进行计数监控。
		reader := io.TeeReader(file, pr)
		io.Copy(part, reader)
	}()

	resp, err := http.Post(baseUrl, multipartWriter.FormDataContentType(), body)
	close(done)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	log.Printf("上传完成，耗时: %v", time.Since(pr.StartTime))
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
