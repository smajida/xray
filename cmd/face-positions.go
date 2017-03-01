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
	"sync/atomic"

	"github.com/lazywei/go-opencv/opencv"
	"github.com/minio/go-cv"
)

var frame uint64

func (v *xrayHandlers) findSimdFaces(currFrame *opencv.Mat) []facePosition {
	// Determine which index to use into array of detect structs
	index := atomic.AddUint64(&frame, 1) % globalDetectParallel

	globalDetectMutex[index].Lock()
	jsonObjects := gocv.DetectObjects(currFrame, globalDetect[index])
	globalDetectMutex[index].Unlock()

	var objInfo ObjectInfo
	if err := json.Unmarshal([]byte(jsonObjects), &objInfo); err != nil {
		panic(err)
	}

	var facePositions []facePosition
	for _, pos := range objInfo.Objects {
		facePositions = append(facePositions, facePosition{
			PT1: Point{
				X: pos.Right,
				Y: pos.Top,
			},
			PT2: Point{
				X: pos.Left,
				Y: pos.Bottom,
			},
			Scalar:    255.0,
			Thickness: 3, // Border thickness defaulted to '3'.
			LineType:  1,
			Shift:     0,
		})
	}
	return facePositions
}
