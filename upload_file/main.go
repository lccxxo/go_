package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

/*
	作用：上传大文件时对上传大小以及百分比进行记录
*/

// progressReader 跟踪上传的大小
type progressReader struct {
	io.Reader
	total     int64
	read      int64
	mu        sync.Mutex
	StartTime time.Time
}

// 实现方法Read，可作为io.Reader
func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	pr.mu.Lock()
	pr.read += int64(n)
	pr.mu.Unlock()
	return
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
	router := gin.Default()

	// 创建上传目录
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		panic(err)
	}

	// 注册路由
	router.POST("/upload", uploadChunkHandler)

	// 启动服务器
	_ = router.Run(":8080")
}

// 文件分块上传接口
func uploadChunkHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()

	pr := &progressReader{
		Reader: src,
		total:  file.Size,
	}

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	// 用来监视是否上传完成
	done := make(chan bool)

	// 用来打印百分比以及关闭channel
	go func() {
		for {
			select {
			case <-ticker.C:
				percent, uploaded := pr.getProgress()
				log.Printf("[Upload Progress] %.2f%% (%d/%d bytes)",
					percent, uploaded, pr.total)
			case <-done:
				percent, uploaded := pr.getProgress()
				log.Printf("[Upload Progress] %.2f%% (%d/%d bytes)",
					percent, uploaded, pr.total)
				return
			}
		}
	}()

	//	保存文件
	dst, err := os.Create("./uploads/" + file.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer dst.Close()

	//	上传文件 pr实现了Read方法可以作为Reader，并且将文件上传大小实时记录
	_, err = io.Copy(dst, pr)
	//	上传完成关闭Done通道 通知channel关闭
	close(done)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "uploaded",
		"filename": file.Filename,
		"size":     file.Size,
	})
}
