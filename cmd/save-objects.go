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
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// sum256 calculate sha256 sum for an input byte array.
func sum256(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

func objectPrefix(data []byte) string {
	sha256Sum := hex.EncodeToString(sum256(data))
	return fmt.Sprintf("%s/%s/%s", sha256Sum[:2], sha256Sum[2:4], sha256Sum[4:])
}

// Uploads image data to configured S3 compatible server using PutObject.
func (v *xrayHandlers) uploadImageData(data []byte) {
	objectName := objectPrefix(data)
	bucketName := globalMinioClntConfig.BucketName()

	// Convert bytes to io.Reader as PutObject expects. Handle error here.
	_, err := v.minioClient.PutObject(bucketName, objectName, bytes.NewReader(data), "image/jpg")
	errorIf(err, "Unable to save image at %s/%s", bucketName, objectName)
}
