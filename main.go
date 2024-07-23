package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	fileQueue  = make(chan string, 10)
	queueMutex = &sync.Mutex{}
	videoDir   = fmt.Sprintf("videos")
	rtmpURL    = "rtmp://a.rtmp.youtube.com/live2/wz26-jb54-6uj9-ujcc-ejxm"
)

func main() {
	router := gin.Default()

	go processQueue()

	router.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("video")
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
			return
		}

		// Generate a unique file path using epoch time
		epochTime := strconv.FormatInt(time.Now().UnixNano(), 10)
		filePath := fmt.Sprintf("%s/%s.mp4", videoDir, epochTime)

		// Create the directory if it doesn't exist
		if err := os.MkdirAll(videoDir, os.ModePerm); err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("directory creation error: %s", err.Error()))
			return
		}

		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}

		// Add the file to the queue for processing
		fileQueue <- filePath

		c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
	})

	router.Run(":8082")
}

// Process the video files from the queue
func processQueue() {
	for filePath := range fileQueue {
		queueMutex.Lock()
		go processVideoFile(filePath)
		queueMutex.Unlock()
	}
}

// Process video file and stream it
func processVideoFile(filePath string) {
	if err := streamToRTMP(filePath); err != nil {
		fmt.Printf("Error streaming to RTMP: %v\n", err)
		return
	}

	// Remove the processed video chunk file
	if err := os.Remove(filePath); err != nil {
		fmt.Printf("Error removing file %s: %v\n", filePath, err)
	}
}

// Stream the video file to RTMP server
func streamToRTMP(filePath string) error {
	fmt.Println("Started Streaming", filePath)
	cmd := exec.Command("ffmpeg", "-re", "-i", filePath, "-c:v", "copy", "-b:v", "2500k", "-c:a", "aac", "-strict", "experimental", "-f", "flv", rtmpURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error executing FFmpeg command: %v", err)
	}
	fmt.Println("Streaming completed successfully.", filePath)
	return nil
}
