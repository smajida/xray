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
	"errors"
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

func (v *xrayHandlers) detectObjects(data []byte) {
	defer func() {
		if r := recover(); r != nil {
			errorIf(r.(error), "Recovered from a panic in detectObjects")
		}
	}()
	img := opencv.DecodeImageMem(data)
	if img == nil {
		errorIf(errors.New("Invalid image input detected"), "Unable to decode incoming image")
		return
	}
	faces := globalHaarCascadeClassifier.DetectObjects(img)
	if len(faces) > 0 {
		var facePositions []facePosition
		for _, value := range faces {
			facePositions = append(facePositions, facePosition{
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
		v.metadataCh <- faceObject{
			Type:      Human,
			Positions: facePositions,
		}
	} else {
		v.metadataCh <- faceObject{
			Type: Unknown,
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
