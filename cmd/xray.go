package xray

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/lazywei/go-opencv/opencv"
)

var haarCasde *opencv.HaarCascade
var upgrader websocket.Upgrader

func init() {
	haarCasde = opencv.LoadHaarClassifierCascade("haarcascade_frontalface_alt.xml")
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	} // use default options
}

type videoRouter struct {
	metadataCh chan []byte // Can be enhanced to something else.
}

type faceType string

const (
	unknownFace faceType = "unknown"
	humanFace   faceType = "human"
	animalFace  faceType = "animal"
)

type faceObject struct {
	Positions []position
	FaceType  faceType
}

type position struct {
	PT1, PT2  opencv.Point
	Color     opencv.Scalar
	Thickness int
	LineType  int
	Shift     int
}

func debugln(v ...interface{}) {
	if os.Getenv("DEBUG") == "1" {
		log.Println(v...)
	}
}

func debugf(format string, v ...interface{}) {
	if os.Getenv("DEBUG") == "1" {
		log.Printf(format, v...)
	}
}

// Send the json data.
func sendData(fo faceObject, metadataCh chan<- []byte) {
	fobytes, err := json.Marshal(&fo)
	if err != nil {
		debugf("Marshalling json error", err)
		return
	}
	metadataCh <- fobytes
}

func (v *videoRouter) detectObjects(data []byte) {
	defer func() {
		if r := recover(); r != nil {
			debugln("Recovered in f", r)
		}
	}()
	img := opencv.DecodeImageMem(data)
	if img == nil {
		debugln("Image is bad.")
		v.metadataCh <- []byte("Image is bad.")
		return
	}
	debugln("Incoming image", img.Channels(), img.Width(), img.Height(), img.ImageSize())
	faces := haarCasde.DetectObjects(img)
	if len(faces) > 0 {
		var positions []position
		for _, value := range faces {
			positions = append(positions, position{
				PT1: opencv.Point{
					X: value.X() + value.Width(),
					Y: value.Y(),
				},
				PT2: opencv.Point{
					X: value.X(),
					Y: value.Y() + value.Height(),
				},
				Color:     opencv.ScalarAll(255.0),
				Thickness: 1,
				LineType:  1,
				Shift:     0,
			})
		}
		debugf("Found humans")
		fo := faceObject{
			FaceType:  humanFace,
			Positions: positions,
		}
		sendData(fo, v.metadataCh)
	} else {
		debugf("No humans found")
		fo := faceObject{
			FaceType: unknownFace,
		}
		sendData(fo, v.metadataCh)
	}
}

func (v *videoRouter) metadata(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error", err)
		return
	}
	debugf("Connected from %s\n", r.Header)
	defer c.Close()
	for {
		mt, data, err := c.ReadMessage()
		if err != nil {
			log.Println("ReadMessage", err)
			break
		}
		debugf("Received message of type %d, with length %d\n", mt, len(data))
		if mt != websocket.BinaryMessage {
			log.Printf("Unrecognized incoming message type %d\n", websocket.TextMessage)
			break
		}
		go v.detectObjects(data)
		if err = c.WriteMessage(websocket.TextMessage, <-v.metadataCh); err != nil {
			log.Println("Error writing to client", err)
		}
	}
}

// Main - X-Ray server entry point.
func Main() {
	log.SetFlags(0)

	v := &videoRouter{
		metadataCh: make(chan []byte),
	}

	http.HandleFunc("/", v.metadata)
	log.Println("Started listening on ws://0.0.0.0:8080")
	log.Fatalln(http.ListenAndServe(":8080", nil))
}
