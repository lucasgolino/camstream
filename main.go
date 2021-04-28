package main

import (
	"fmt"
	"image"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"gocv.io/x/gocv"
)

var (
	err      error
	webcam   *gocv.VideoCapture
)

var buffer = make(map[int][]byte)
var frame []byte
var mutex = &sync.Mutex{}

func main() {

	fmt.Println("[INFO] RTSP Stream: " + os.Args[1])
	webcam, err = gocv.VideoCaptureFile(os.Args[1])

	if err != nil {
		fmt.Printf("Error opening capture device: \n")
		return
	}

	defer webcam.Close()

	// start capturing
	go FrameRoutine()

	fmt.Println("[INFO] Capturing")

	// start http server
	http.HandleFunc("/video_feed", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		data := ""
		for {
			mutex.Lock()
			data = "--frame\r\n  Content-Type: image/jpeg\r\n\r\n" + string(frame) + "\r\n\r\n"
			mutex.Unlock()
			time.Sleep(33 * time.Millisecond)
			w.Write([]byte(data))
		}
	})

	http.HandleFunc("/video_snap", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(frame)
	})

	log.Fatal(http.ListenAndServe("0.0.0.0:8888", nil))
}

func FrameRoutine() {
	mat := gocv.NewMat()
	defer mat.Close()

	for {
		if ok := webcam.Read(&mat); !ok {
			fmt.Printf("Device closed\n")
			return
		}

		if mat.Empty() {
			continue
		}

		gocv.Resize(mat, &mat, image.Point{}, float64(0.5), float64(0.5), 0)
		frame, _ = gocv.IMEncode(".jpg", mat)
	}
}