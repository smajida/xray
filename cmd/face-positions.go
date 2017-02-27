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

import "github.com/lazywei/go-opencv/opencv"

func getFacePositions(faces []*opencv.Rect) (facePositions []facePosition, faceFound bool) {
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
			Thickness: 3, // Border thickness defaulted to '3'.
			LineType:  1,
			Shift:     0,
		})
	}

	return facePositions, len(facePositions) > 0
}

func (v *xrayHandlers) findFaces(currFrame *opencv.IplImage) (faces []*opencv.Rect) {
	return globalHaarCascadeClassifier.DetectObjects(currFrame)
}
