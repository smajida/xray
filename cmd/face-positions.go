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
	"github.com/fwessels/go-cv"
	"github.com/lazywei/go-opencv/opencv"
	"sync/atomic"
)

func getFacePositions(faces []*opencv.Rect) (facePositions []facePosition) {
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
			Scalar:    255.0,
			Thickness: 3, // Border thickness defaulted to '3'.
			LineType:  1,
			Shift:     0,
		})
	}
	return facePositions
}

func (v *xrayHandlers) findFaces(currFrame *opencv.IplImage) (faces []*opencv.Rect) {
	// Default haar cascade classifier used for detecting faces.
	globalHaarCascadeClassifier := opencv.LoadHaarClassifierCascade("haarcascade_frontalface_alt.xml")
	return globalHaarCascadeClassifier.DetectObjects(currFrame)
}

type Object struct {
	Objects []Rectangle
}

type Rectangle struct {
	Top    int
	Left   int
	Bottom int
	Right  int
}

var frame uint64 = 0

func (v *xrayHandlers) findSimdFaces(currFrame *opencv.Mat) []*opencv.Rect {

	// Determine which index to use into array of detect structs
	index := atomic.AddUint64(&frame, 1) % globalDetectParallel

	globalDetectMutex[index].Lock()
	jsonObjects := gocv.DetectObjects(currFrame, globalDetect[index])
	globalDetectMutex[index].Unlock()

	var objs Object
	json.Unmarshal([]byte(jsonObjects), &objs)

	var faces []*opencv.Rect

	for _, obj := range objs.Objects {
		rect := new(opencv.Rect)
		rect.Init(obj.Left, obj.Top, obj.Right-obj.Left, obj.Bottom-obj.Top)
		faces = append(faces, rect)
	}

	return faces
}
