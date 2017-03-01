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

import "math"

// Face type..
type faceType string

// List of face types.
const (
	Unknown      faceType = "unknown"
	Human        faceType = "human"
	HumanToddler faceType = "human-toddler"
)

// Face object contains all the data
// to be sent back to the client.
type faceObject struct {
	Positions []facePosition
	Type      faceType
	Display   bool
	Zoom      int
}

// Represents the spacial rectangular co-ordinates
// of face position in a frame.
type facePosition struct {
	// Co-ordinates for drawing rectangle.
	PT1, PT2 Point
	Scalar   float64
	// Overally border thickness of the rectangle.
	Thickness int
	// Line type of the overlay rectangle.
	LineType int
	Shift    int
}

type Point struct {
	X int
	Y int
}

func (p Point) Add(p2 Point) Point {
	p.X += p2.X
	p.Y += p2.Y
	return p
}

func (p Point) Sub(p2 Point) Point {
	p.X -= p2.X
	p.Y -= p2.Y
	return p
}

func (p Point) Radius() float64 {
	return math.Sqrt(p.RadiusSq())
}

func (p Point) RadiusSq() float64 {
	return float64(p.X*p.X + p.Y*p.Y)
}

func (p Point) Angle() float64 {
	return math.Atan2(float64(p.Y), float64(p.X))
}

// ObjectInfo - represents face positions.
type ObjectInfo struct {
	Positions []Rectangle
}

type Rectangle struct {
	Top    int
	Left   int
	Bottom int
	Right  int
}
