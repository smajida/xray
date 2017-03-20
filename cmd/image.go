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
	"image"
	"math"
	"strconv"
)

type frameStruct struct {
	ID        string `json:"id"`
	Format    string `json:"format"`
	Width     string `json:"width"`
	Height    string `json:"height"`
	Rotation  string `json:"rotation"`
	Timestamp string `json:"timestamp"`
}

type faceStruct struct {
	ID           string      `json:"id"`
	EulerY       string      `json:"eulerY"`
	EulerZ       string      `json:"eulerZ"`
	Height       string      `json:"height"`
	Width        string      `json:"width"`
	LeftEyeOpen  string      `json:"leftEyeOpen"`
	RightEyeOpen string      `json:"rightEyeOpen"`
	Smiling      string      `json:"similing"`
	FacePt1      pointStruct `json:"facePt1"`
	FacePt2      pointStruct `json:"facePt2"`
}

type pointStruct struct {
	X string `json:"x"`
	Y string `json:"y"`
}

type frameRecord struct {
	Frame frameStruct  `json:"frame"`
	Faces []faceStruct `json:"faces"`
}

// Extracts full frame rectangle from the incoming frame record.
func (fr *frameRecord) GetFullFrameRect() (image.Rectangle, int, error) {
	width, err := strconv.Atoi(fr.Frame.Width)
	if err != nil {
		return image.Rectangle{}, 0, err
	}
	height, err := strconv.Atoi(fr.Frame.Width)
	if err != nil {
		return image.Rectangle{}, 0, err
	}

	frameID, err := strconv.Atoi(fr.Frame.ID)
	if err != nil {
		return image.Rectangle{}, 0, err
	}

	return image.Rectangle{image.Point{}, image.Point{X: width, Y: height}}, frameID, nil
}

// Extracts all the face rectangles from the incoming frame record.
func (fr *frameRecord) GetFaceRectangles() ([]image.Rectangle, error) {
	var faces []image.Rectangle
	for _, face := range fr.Faces {
		x1, err := strconv.ParseFloat(face.FacePt1.X, 64)
		if err != nil {
			return []image.Rectangle{}, err
		}
		y1, err := strconv.ParseFloat(face.FacePt1.Y, 64)
		if err != nil {
			return []image.Rectangle{}, err
		}
		x2, err := strconv.ParseFloat(face.FacePt2.X, 64)
		if err != nil {
			return []image.Rectangle{}, err
		}
		y2, err := strconv.ParseFloat(face.FacePt2.Y, 64)
		if err != nil {
			return []image.Rectangle{}, err
		}
		faces = append(faces, image.Rectangle{
			image.Point{X: int(x1), Y: int(y1)}, image.Point{X: int(x2), Y: int(y2)},
		})
	}

	return faces, nil
}

// Rectangle represents custom rectangle implementation, provides
// additional methods for calculating threshold factors.
type Rectangle image.Rectangle

// In reports whether every point in r is in s, under a given threshold factor.
func (r Rectangle) In(s image.Rectangle, thresholdFactor int) bool {
	// Note that r.Max is an exclusive bound for r, so that r.In(s)
	// does not require that r.Max.In(s).
	return s.Min.X-r.Min.X <= thresholdFactor && r.Max.X-s.Max.X <= thresholdFactor &&
		s.Min.Y-r.Min.Y <= thresholdFactor && r.Max.Y-s.Max.Y <= thresholdFactor
}

// Algorithm used here is pretty simple union of face rectangles is verified
// if it is sufficiently visible under the incoming frame, if the image is
// sufficiently visible then the optimal zoom factor is calculated based on
// the available room for the individual maximal and minimum points of
// all the faces in a given frame.
//
// Currently the supported threshold point difference are
// 100 - which would yield '0' zoom factor.
// 200 - which would yield '1' zoom factor.
// 300 - which would yield '2' zoom factor.
// ... To support more area.
func calculateOptimalZoomFactor(faces []image.Rectangle, rect image.Rectangle) int {
	var final image.Rectangle
	for _, rect := range faces {
		final = final.Union(rect)
	}

	if final.Empty() {
		return -1
	}

	r := Rectangle{final.Max, final.Min}

	for i, ifactor := range []int{100, 200, 300} {
		if r.In(rect, ifactor) {
			return i
		}
	}

	return -1
}

// Point represents - 2D points specified by its coordinates x and y.
type Point image.Point

// Radius - calculate the radius between the points.
func (p Point) Radius() float64 {
	return math.Sqrt(p.RadiusSq())
}

// RadiusSq - calculate raidus square X^2+Y^2
func (p Point) RadiusSq() float64 {
	return float64(p.X*p.X + p.Y*p.Y)
}

// Angle - calculate arc tangent of Y/X
func (p Point) Angle() float64 {
	return math.Atan2(float64(p.Y), float64(p.X))
}
