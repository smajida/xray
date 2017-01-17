package xray

import (
	"bytes"
	"log"
	"net/http"

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

func (v *videoRouter) detectObjects(data []byte) {
	a := bytes.Index(data, []byte("\xff\xd8"))
	b := bytes.Index(data, []byte("\xff\xd9"))
	if a == -1 || b == -1 {
		v.metadataCh <- "Waiting for MJPEG"
		return
	}
	if a > b+2 {
		v.metadataCh <- "Waiting for MJPEG"
		return
	}
	log.Println("Found MJPEG", a, b, len(data))
	jpg := data[a : b+2]
	faces := haarCasde.DetectObjects(opencv.DecodeImageMem(jpg))
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
	log.Printf("Connected from %s\n", r.Header)
	defer c.Close()
	for {
		mt, data, err := c.ReadMessage()
		if err != nil {
			log.Println("ReadMessage", err)
			break
		}
		log.Printf("Received message of type %d, with length %d\n", mt, len(data))
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
