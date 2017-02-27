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
	"bytes"
	"encoding/json"
	"net/http"
	"sync"

	router "github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/lazywei/go-opencv/opencv"
)

type xrayHandlers struct {
	sync.RWMutex

	// Used for calculating difference.
	prevSR sensorRecord

	// Represents client response channel, sends client data.
	clntRespCh chan interface{}

	// Used for upgrading the incoming HTTP
	// wconnection into a websocket wconnection.
	upgrader websocket.Upgrader
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

	faces := v.findFaces(img)
	if len(faces) == 0 {
		fo := faceObject{
			Type:    Unknown,
			Display: true,
			Zoom:    0,
		}
		v.writeClntData(fo)
		return
	}

	// Detected faces, decode their positions.
	facePositions := getFacePositions(faces)
	faceFound := len(facePositions) > 0
	fo := faceObject{
		Type:      Human,
		Positions: facePositions,
		Display:   true || faceFound,
	}

	// Zooming happens relative on Android if faces are detected.
	if faceFound {
		switch len(facePositions) {
		case 1:
			// For single face detection zoom in.
			fo.Zoom = 1
		default:
			// For more than 1 Zoom out for more coverage.
			fo.Zoom = -1
		}
	}

	// Send the data to client.
	v.writeClntData(fo)
}

// Detects motion based on sensor difference.
func (v *xrayHandlers) detectMotion(data []byte) {
	defer func() {
		if r := recover(); r != nil {
			errorIf(r.(error), "Recovered from a panic in detectMotion")
		}
	}()

	if !bytes.Contains(data, []byte("sensorName")) {
		return
	}

	var sr sensorRecord
	if err := json.Unmarshal(data, &sr); err != nil {
		errorIf(err, "Unable to extract sensor record %s", string(data))
		return
	}

	fo := faceObject{
		Type:    Unknown,
		Display: true || v.shouldDisplayCamera(sr),
		Zoom:    0,
	}

	v.writeClntData(fo)
	v.persistCurrentSensor(sr)
}

// Detect detects metadata about the incoming data.
func (v *xrayHandlers) Detect(w http.ResponseWriter, r *http.Request) {
	wconn, err := v.upgrader.Upgrade(w, r, nil)
	if err != nil {
		errorIf(err, "Unable to perform websocket upgrade the request.")
		return
	}
	wc := wConn{wconn}
	defer wc.Close()

	// Waiting on incoming reads.
	for {
		mt, data, err := wc.ReadMessage()
		if err != nil {
			errorIf(err, "Unable to read incoming binary message.")
			break
		}

		if len(data) == 0 {
			continue
		}

		// Support if client sent a text message, most
		// probably its sensor or location metadata.
		if mt == websocket.TextMessage {
			go v.detectMotion(data)
			wc.WriteMessage(mt, v.clntRespCh)
			continue
		}

		if mt == websocket.BinaryMessage {
			go v.detectObjects(data)
			wc.WriteMessage(mt, v.clntRespCh)
		}
	}
}

// Initialize a new xray handlers.
func newXRayHandlers() *xrayHandlers {
	return &xrayHandlers{
		clntRespCh: make(chan interface{}, 1000),
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
	xrayRouter.Methods("GET").HandlerFunc(xray.Detect)
}
