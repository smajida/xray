/*
 * Copyright (c) 2017 Minio, Inc. <https://www.minio.io>
 *
 * This file is part of Xray.
 *
 * Xray is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package cmd

import (
	"encoding/json"
	"net/http"

	router "github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/lazywei/go-opencv/opencv"
)

type xrayHandlers struct {
	// Contains the data to be sent
	// back over the wire.
	metadataCh chan faceObject

	// Used for upgrading the incoming HTTP
	// wconnection into a websocket wconnection.
	upgrader websocket.Upgrader
}

// Find contours in incoming image.
// @Deprecated
func findContours(img *opencv.IplImage, pos int) *opencv.Seq {
	w := img.Width()
	h := img.Height()

	// Create the output image
	cedge := opencv.CreateImage(w, h, opencv.IPL_DEPTH_8U, 3)
	defer cedge.Release()

	// Convert to grayscale
	gray := opencv.CreateImage(w, h, opencv.IPL_DEPTH_8U, 1)
	edge := opencv.CreateImage(w, h, opencv.IPL_DEPTH_8U, 1)
	defer gray.Release()
	defer edge.Release()

	opencv.CvtColor(img, gray, opencv.CV_BGR2GRAY)

	opencv.Smooth(gray, edge, opencv.CV_BLUR, 3, 3, 0, 0)
	opencv.Not(gray, edge)

	// Run the edge detector on grayscale
	opencv.Canny(gray, edge, float64(pos), float64(pos*3), 3)

	opencv.Zero(cedge)
	// copy edge points
	opencv.Copy(img, cedge, edge)

	return edge.FindContours(opencv.CV_RETR_TREE, opencv.CV_CHAIN_APPROX_SIMPLE, opencv.Point{0, 0})
}

// Detects face objects on incoming data.
func (v *xrayHandlers) detectObjects(data []byte) {
	defer func() {
		if r := recover(); r != nil {
			errorIf(r.(error), "Recovered from a panic in detectObjects")
		}
	}()
	img := opencv.DecodeImageMem(data)
	if img == nil {
		errorIf(errInvalidImage, "Unable to decode incoming image")
		return
	}
	faces := globalHaarCascadeClassifier.DetectObjects(img)
	if len(faces) > 0 {
		var facePositions []facePosition
		for _, value := range faces {
			if value.X() == 0 || value.Y() == 0 {
				continue
			}
			facePositions = append(facePositions, facePosition{
				PT1: opencv.Point{
					X: value.X() + value.Width(),
					Y: value.Y(),
				},
				PT2: opencv.Point{
					X: value.X(),
					Y: value.Y() + value.Height(),
				},
				Scalar:    255.0,
				Thickness: 3, // Border thickness defaulted.
				LineType:  1,
				Shift:     0,
			})
		}
		seq := findContours(img, 2)
		v.metadataCh <- faceObject{
			Type:      Human,
			Contours:  seq,
			Positions: facePositions,
			Display:   true,
			// Dummy value needs to be addressed in future.
			Zoom: 1,
		}
		seq.Release()
		img.Release()
	} else {
		v.metadataCh <- faceObject{
			Type:    Unknown,
			Display: false,
		}
	}
}

// Write json data back.
func writeFaceObject(wconn *websocket.Conn, metadataCh <-chan faceObject) {
	fo := <-metadataCh
	if fo.Type == Unknown {
		// Not writing to client if face is unknown.
		return
	}
	fobytes, err := json.Marshal(&fo)
	if err != nil {
		errorIf(err, "Unable to marshal %#v into json.", fo)
		return
	}
	if err = wconn.WriteMessage(websocket.TextMessage, fobytes); err != nil {
		errorIf(err, "Unable to write to client.")
		return
	}
}

// DetectObject reads the incoming data.
func (v *xrayHandlers) DetectObject(w http.ResponseWriter, r *http.Request) {
	wconn, err := v.upgrader.Upgrade(w, r, nil)
	if err != nil {
		errorIf(err, "Unable to perform websocket upgrade the request.")
		return
	}
	defer wconn.Close()
	for {
		mt, data, err := wconn.ReadMessage()
		if err != nil {
			errorIf(err, "Unable to read incoming binary message.")
			break
		}
		if mt != websocket.BinaryMessage {
			errorIf(err, "Unable to recognize incoming message type %d\n", mt)
			continue
		}
		go v.detectObjects(data)
		writeFaceObject(wconn, v.metadataCh)
	}
}

// Initialize a new xray handlers.
func newXRayHandlers() *xrayHandlers {
	return &xrayHandlers{
		metadataCh: make(chan faceObject),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}, // use default options
	}
}

// Configure xray handler.
func configureXrayHandler(mux *router.Router) http.Handler {
	registerXRayRouter(mux)

	// Register additional routers if any.
	return mux
}

// Register xray router.
func registerXRayRouter(mux *router.Router) {
	// Initialize xray handlers.
	xray := newXRayHandlers()

	// xray Router
	xrayRouter := mux.NewRoute().PathPrefix("/").Subrouter()

	// Currently there is only one handler.
	xrayRouter.Methods("GET").HandlerFunc(xray.DetectObject)
}
