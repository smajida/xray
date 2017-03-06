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
)

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
func calculateOptimalZoomFactor(faces []facePosition, rect image.Rectangle) int {
	var faceRectangles []image.Rectangle
	for _, facePos := range faces {
		faceRectangles = append(faceRectangles, image.Rectangle{
			image.Point(facePos.PT1), image.Point(facePos.PT2),
		})
	}

	var final image.Rectangle
	for _, rect := range faceRectangles {
		final.Union(rect)
	}

	if final.Empty() {
		return -1
	}

	for i, ifactor := range []int{100, 200, 300} {
		r := Rectangle{final.Max, final.Min}
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
