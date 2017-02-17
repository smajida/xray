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

// Find contours in incoming image. @Deprecated
func findContours(img *opencv.IplImage, pos int) *opencv.Seq {
	w := img.Width()
	h := img.Height()

	// Create the output image
	cedge := opencv.CreateImage(w, h, opencv.IPL_DEPTH_8U, 3)
	defer cedge.Release()

	// Convert to grayscale
	gray := opencv.CreateImage(w, h, opencv.IPL_DEPTH_8U, 1)
	edge := opencv.CreateImage(w, h, opencv.IPL_DEPTH_8U, 1)
	defer gray.Release()
	defer edge.Release()

	opencv.CvtColor(img, gray, opencv.CV_BGR2GRAY)

	opencv.Smooth(gray, edge, opencv.CV_BLUR, 3, 3, 0, 0)
	opencv.Not(gray, edge)

	// Run the edge detector on grayscale
	opencv.Canny(gray, edge, float64(pos), float64(pos*3), 3)

	opencv.Zero(cedge)
	// copy edge points
	opencv.Copy(img, cedge, edge)

	return edge.FindContours(opencv.CV_RETR_TREE, opencv.CV_CHAIN_APPROX_SIMPLE, opencv.Point{0, 0})
}
