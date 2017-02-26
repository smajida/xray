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

// Returns if there is motion with in the globally accepted threshold.
func isThereMotion(motionFrame, origFrame *opencv.IplImage) bool {
	// Detect motion in window
	xStart, xStop := 10, origFrame.GetMat().Cols()-11
	yStart, yStop := 10, origFrame.GetMat().Rows()-11

	minX := motionFrame.Width()
	minY := motionFrame.Height()

	_, stdDev := motionFrame.MeanStdDev()
	if int(stdDev.Val()[0]) > globalMaxDeviation {
		return false
	} // else < globalMaxDeviation

	var (
		numberOfChanges int
		maxX            int
		maxY            int
	)

	for j := yStart; j < yStop; j += 2 {
		for i := xStart; i < xStop; i += 2 {
			if int(motionFrame.Get2D(i, j).Val()[0]) == 255 {
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

	return numberOfChanges > globalThereIsMotion
}

type motionWindowThreshold struct {
	// Detect motion in following window
	xStart, xStop int
	yStart, yStop int
}

const (
	// If more than 'thereIsMotion' pixels are changed, we say there is motion.
	globalThereIsMotion = 5

	// Maximum deviation of the image, the higher the value, the more motion is allowed
	globalMaxDeviation = 20
)

func detectMovingFrames(currFrame, nextFrame *opencv.IplImage) bool {
	// create the needed frames
	w := currFrame.Width()
	h := currFrame.Height()

	cfg := opencv.CreateImage(w, h, opencv.IPL_DEPTH_8U, 1)
	nfg := opencv.CreateImage(w, h, opencv.IPL_DEPTH_8U, 1)

	// convert to grayscale
	opencv.CvtColor(currFrame, cfg, opencv.CV_BGR2GRAY)
	opencv.CvtColor(currFrame, nfg, opencv.CV_BGR2GRAY)

	kernelErode := opencv.CreateStructuringElement(2, 2, 1, 1, opencv.CV_MORPH_RECT)
	defer kernelErode.ReleaseElement()

	pfg := cfg
	cfg = nfg

	nfg = opencv.CreateImage(w, h, opencv.IPL_DEPTH_8U, 1)
	opencv.CvtColor(nextFrame, nfg, opencv.CV_BGR2GRAY)

	motion := diffImg(pfg, cfg, nfg)
	defer pfg.Release()
	defer motion.Release()

	opencv.Threshold(motion, motion, float64(10), 255, opencv.CV_THRESH_BINARY)
	opencv.Erode(motion, motion, kernelErode, 1)

	return isThereMotion(motion, pfg)
}

// Detects if one should display camera.
func (v *xrayHandlers) shouldDisplayCamera(currFrame *opencv.IplImage) {
	var display bool

	if v.prevFrame != nil && currFrame != nil {
		display = detectMovingFrames(v.prevFrame, currFrame)
	}

	fo := faceObject{
		Type:    Unknown,
		Display: display,
		Zoom:    -1,
	}

	v.writeClntData(fo)
}
