package xray

import (
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
	metadataCh chan string // Can be enhanced to something else.
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

func (v *videoRouter) detectObjects(data []byte) {
	defer func() {
		if r := recover(); r != nil {
			debugln("Recovered in f", r)
		}
	}()
	img := opencv.DecodeImageMem(data)
	if img == nil {
		debugln("Image is bad.")
		v.metadataCh <- "Image is bad."
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
		debugf("Found humans at positions %#v\n", positions)
		v.metadataCh <- "Humans"
	} else {
		v.metadataCh <- "No humans found"
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
		if err = c.WriteMessage(websocket.TextMessage, []byte(<-v.metadataCh)); err != nil {
			log.Println("Error writing to client", err)
		}
	}
}

// Main - X-Ray server entry point.
func Main() {
	log.SetFlags(0)

	v := &videoRouter{
		metadataCh: make(chan string),
	}

	http.HandleFunc("/", v.metadata)
	log.Println("Started listening on ws://0.0.0.0:8080")
	log.Fatalln(http.ListenAndServe(":8080", nil))
}
