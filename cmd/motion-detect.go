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

var (
	maxDevitaion int
)

func diffImg(t0, t1, t2 *opencv.IplImage) (diff *opencv.IplImage) {
	w := t0.Width()
	h := t0.Height()

	d1 := opencv.CreateImage(w, h, opencv.IPL_DEPTH_8U, 1)
	defer d1.Release()

	d2 := opencv.CreateImage(w, h, opencv.IPL_DEPTH_8U, 1)
	defer d2.Release()

	diff = opencv.CreateImage(w, h, opencv.IPL_DEPTH_8U, 1)

	opencv.AbsDiff(t2, t1, d1)
	opencv.AbsDiff(t1, t0, d2)
	opencv.And(d1, d2, diff)

	return

}

func detectMotion(motion, result *opencv.IplImage, xStart, xStop, yStart, yStop, maxDevitaion int) int {

	_, stdDev := motion.MeanStdDev()

	if int(stdDev.Val()[0]) < maxDevitaion {
		numberOfChanges, maxX, maxY := 0, 0, 0
		minX := motion.Width()
		minY := motion.Height()

		for j := yStart; j < yStop; j += 2 {
			for i := xStart; i < xStop; i += 2 {
				if int(motion.Get2D(i, j).Val()[0]) == 255 {
					numberOfChanges++
					if minX > i {
						minX = i
					}
					if maxX < i {
						maxX = i
					}
					if minY > j {
						minY = j
					}
					if maxY < j {
						maxY = i
					}
				}
			}
		}

		if numberOfChanges > 0 {
			if minX-10 > 0 {
				minX -= 10
			}
			if minY-10 > 0 {
				minY -= 10
			}
			if maxX+10 < result.Width()-1 {
				maxX += 10
			}
			if maxY+10 < result.Height()-1 {
				maxY += 10
			}

			var pt1 opencv.Point
			var pt2 opencv.Point

			pt1.X = minX
			pt1.Y = minY

			pt2.X = maxX
			pt2.Y = maxY
		}
		return numberOfChanges

	}
	return 0
}

func detectMovingFrames(frame, img *opencv.IplImage) bool {
	// create the needed frames
	w := frame.Width()
	h := frame.Height()

	currentFrame := opencv.CreateImage(w, h, opencv.IPL_DEPTH_8U, 1)
	nextFrame := opencv.CreateImage(w, h, opencv.IPL_DEPTH_8U, 1)

	// convert to grayscale
	opencv.CvtColor(frame, currentFrame, opencv.CV_BGR2GRAY)
	opencv.CvtColor(frame, nextFrame, opencv.CV_BGR2GRAY)

	var (
		// Detect motion in window
		xStart = 10
		xStop  = currentFrame.GetMat().Cols() - 11
		yStart = 10
		yStop  = currentFrame.GetMat().Rows() - 11

		// If more than 'thereIsMotion' pixels are changed, we say there is motion.
		thereIsMotion = 5

		// Maximum deviation of the image, the higher the value, the more motion is allowed
		maxDeviation = 20
	)

	kernelErode := opencv.CreateStructuringElement(2, 2, 1, 1, opencv.CV_MORPH_RECT)
	defer kernelErode.ReleaseElement()

	prevFrame := currentFrame
	currentFrame = nextFrame
	result := img.Clone()
	nextFrame = opencv.CreateImage(w, h, opencv.IPL_DEPTH_8U, 1)
	opencv.CvtColor(img, nextFrame, opencv.CV_BGR2GRAY)

	motion := diffImg(prevFrame, currentFrame, nextFrame)
	// // motion := nextFrame.Clone()
	opencv.Threshold(motion, motion, float64(10), 255, opencv.CV_THRESH_BINARY)
	opencv.Erode(motion, motion, kernelErode, 1)

	numberOfChanges := detectMotion(motion, result, xStart, xStop, yStart, yStop, maxDeviation)
	prevFrame.Release()
	result.Release()
	motion.Release()

	return numberOfChanges > thereIsMotion
}
