package cmd

import (
	"encoding/json"
	"fmt"
	"image"
	"testing"
	"time"
)

var frame = 0
var timestamp = 99

const frameWidth = 960
const frameHeight = 720

func rectFromCenter(pt image.Point, size int) image.Rectangle {

	return image.Rect(pt.X-size, pt.Y-size, pt.X+size, pt.Y+size)
}

func getJson(rects []image.Rectangle) string {

	frame++
	timestamp++

	jsonFaces := ""
	for id, rect := range rects {
		jsonFaces += fmt.Sprintf(`{ "id": "%d", "eulerY": "0.0", "eulerZ": "0.0", "width": "%d", "height": "%d", "leftEyeOpen": "-1.0", "rightEyeOpen": "-1.0", "similing": "0.0", "facePt1": { "x": "%d", "y": "%d"  }, "facePt2": {  "x": "%d",  "y": "%d" } }`,
			id+1, rect.Dx(), rect.Dy(), rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y)
		if id < len(rects)-1 {
			jsonFaces += ",\n             "
		}
	}

	return fmt.Sprintf(`{ "frame": { "id": "%d", "format": "17", "width": "%d", "height": "%d", "rotation": "2", "timestamp": "%d" }, "faces": [ %s ] }`, frame, frameWidth, frameHeight, timestamp, jsonFaces)
}

func TestMotion(t *testing.T) {

	mr := MotionRecorder{}

	for size := 25; size < 20000; size += 1 {

		time.Sleep(time.Millisecond * 20)

		rect := rectFromCenter(image.Point{X: frameWidth/2 + size, Y: frameHeight / 2}, 25)
		jsontext := getJson([]image.Rectangle{rect})

		var fr frameRecord
		if err := json.Unmarshal([]byte(jsontext), &fr); err != nil {
			panic("failed to unmarshal JSON")
		}

		mr.Append(&fr)
		if mr.DetectMotion() {
			fmt.Println("SNAPSHOT at threshold", mr.Threshold())
		}
	}
}
