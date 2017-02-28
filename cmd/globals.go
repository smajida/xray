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
	"os"
	"sync"
	"unsafe"

	"github.com/fwessels/go-cv"
)

func init() {
	for i := 0; i < globalDetectParallel; i++ {
		globalDetect[i] = gocv.DetectInitialize("haar_face_0.xml")
	}

}

// Global constants for Xray.
const (
	minGoVersion         = ">= 1.7" // Xray requires at least Go v1.7
	globalDetectParallel = 8
)

var (
	globalXrayCertFile = "/etc/ssl/public.crt"
	globalXrayKeyFile  = "/etc/ssl/private.key"
	globalDebug        = os.Getenv("DEBUG") != ""

	globalUseGoCV     = true
	globalDetectMutex [globalDetectParallel]sync.Mutex
	globalDetect      [globalDetectParallel]unsafe.Pointer

	// List of all other classifiers.
)
