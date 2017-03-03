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
	"github.com/minio/minio-go"
	router "github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/minio/go-cv"
)

type xrayHandlers struct {
	sync.RWMutex

	// Object Storage handler.
	minioClient *minio.Client

	// Used for calculating difference.
	prevSR sensorRecord

	// Represents client response channel, sends client data.
	clntRespCh chan interface{}

	// Display memory channels.
	displayCh, displayRecvCh chan bool

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

	img, err := gocv.DecodeImageMem(data)
	if err != nil {
		errorIf(err, "Unable to decode incoming image")
		v.displayCh <- false
		fo := faceObject{
			Type:    Unknown,
			Display: <-v.displayRecvCh,
			Zoom:    0,
		}
		v.clntRespCh <- fo
		return
	}

	// Detected faces, decode their positions.
	faces := v.findSimdFaces(img)
	if len(faces) == 0 {
		v.displayCh <- false
		fo := faceObject{
			Type:    Unknown,
			Display: <-v.displayRecvCh,
			Zoom:    0,
		}
		v.clntRespCh <- fo
		return
	}
	// Save frames to object storage server whenever faces are detected.
	go v.saveObjects(data)

	v.displayCh <- len(faces) > 0
	fo := faceObject{
		Type:      Human,
		Positions: faces,
		Display:   <-v.displayRecvCh,
	}

	// Zooming happens relative on Android if faces are detected.
	if fo.Display {
		switch len(faces) {
		case 1:
			// For single face detection zoom in.
			fo.Zoom = 1
		default:
			// For more than 1 Zoom out for more coverage.
			fo.Zoom = -1
		}
	}

	// Send the data to client.
	v.clntRespCh <- fo
}

// Detects motion based on sensor difference.
func (v *xrayHandlers) detectMotion(data []byte) {
	defer func() {
		if r := recover(); r != nil {
			errorIf(r.(error), "Recovered from a panic in detectMotion")
		}
	}()

	var sr sensorRecord
	if err := json.Unmarshal(data, &sr); err != nil {
		v.displayCh <- false
		fo := faceObject{
			Type:    Unknown,
			Display: <-v.displayRecvCh,
			Zoom:    0,
		}
		v.clntRespCh <- fo
		errorIf(err, "Unable to extract sensor record %s", string(data))
		return
	}

	v.displayCh <- v.shouldDisplayCamera(sr)
	fo := faceObject{
		Type:    Unknown,
		Display: <-v.displayRecvCh,
		Zoom:    0,
	}

	v.clntRespCh <- fo
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
			// Ignore all other forms of incoming data.
			if !bytes.Contains(data, []byte("sensorName")) {
				continue
			}
			go v.detectMotion(data)
		} else if mt == websocket.BinaryMessage {
			go v.detectObjects(data)
		}
		wc.WriteMessage(websocket.TextMessage, v.clntRespCh)
	}
}

// Initialize a new xray handlers.
func newXRayHandlers(clnt *minio.Client) *xrayHandlers {
	displayCh := make(chan bool)

	
	return &xrayHandlers{
		clntRespCh:    make(chan interface{}, 15000),
		displayCh:     displayCh,
		minioClient:   clnt,
		displayRecvCh: displayMemoryRoutine(displayCh),
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

// Create a minio client to play.minio.io and make a bucket.
func newMinioClient() (*minio.Client, error) {
    endpoint := "play.minio.io:9000"
    accessKeyID := "Q3AM3UQ867SPQQA43P2F"
    secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
    useSSL := true
    
    // Initialize minio client object.
    minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
    if err != nil {
    	return nil, err
    }
    
    // Make a new bucked called alice.
    bucketName := "alice"
    location := "us-east-1"

 	// Check to see if we already own this bucket (which happens if you run this twice)
    exists, err := minioClient.BucketExists(bucketName)
    if err != nil {
    	return nil, err
    }
    if !exists {
    	  err = minioClient.MakeBucket(bucketName, location)
    	  if err != nil {
    	  		return nil, err
    	  }
    }
    return minioClient, nil


}
// Register xray router.
func registerXRayRouter(mux *router.Router) {
	// 
	clnt, err := newMinioClient()
	if err != nil {
		panic(err)
	}

	// Initialize xray handlers.
	xray := newXRayHandlers(clnt)

	// xray Router
	xrayRouter := mux.NewRoute().PathPrefix("/").Subrouter()

	// Currently there is only one handler.
	xrayRouter.Methods("GET").HandlerFunc(xray.Detect)
}
