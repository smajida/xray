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

// Point represents - 2D points specified by its coordinates x and y.
type Point struct {
	image.Point
}

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
