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

// Face type..
type faceType string

// List of face types.
const (
	Unknown      faceType = "unknown"
	Human        faceType = "human"
	HumanToddler faceType = "human-toddler"
)

// XRayDetectResult - represents the computed
// metadata by xray server for the incoming image
// data. Includes face positions, face type, optimal
// zoom factor, presignedPOST url for the client to
// save to server.
type XRayDetectResult struct {
	// TODO needs to add frame id.

	// Collection of various faces in the incoming image.
	Positions []FacePosition

	// Type of face, currently only supports "human"
	Type faceType

	// Should the camera turn itself on.
	Display bool

	// Optimal zoom factor for the camera.
	Zoom int

	// Presigned information if any for client
	// to start upload the frames..
	Presigned *presignedPOST
}

// FacePosition Represents the 2D rectangular
// co-ordinates of face position detected from
// the incoming frame.
type FacePosition struct {
	// Co-ordinates for drawing rectangle.
	PT1, PT2 Point
	Scalar   float64

	// Overally border thickness of the rectangle.
	Thickness int

	// Line type of the overlay rectangle.
	LineType int
	Shift    int
}
