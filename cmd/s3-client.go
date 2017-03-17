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
	"fmt"
	"os"
	"strconv"
	"time"

	minio "github.com/minio/minio-go"
)

func getObjectPrefix() string {
	uid := fmt.Sprintf("%x", time.Now().UTC().UnixNano())
	return fmt.Sprintf("%s/%s/%s", uid[:2], uid[2:4], uid[4:])
}

type minioConfig struct{}

func (c minioConfig) Endpoint() string {
	ep := os.Getenv("S3_ENDPOINT")
	if ep == "" {
		// Default
		ep = "play.minio.io:9000"
	}
	return ep
}

func (c minioConfig) AccessKey() string {
	ak := os.Getenv("ACCESS_KEY")
	if ak == "" {
		// Default
		ak = "Q3AM3UQ867SPQQA43P2F"
	}
	return ak
}

func (c minioConfig) SecretKey() string {
	sk := os.Getenv("SECRET_KEY")
	if sk == "" {
		// Default
		sk = "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	}
	return sk
}

func (c minioConfig) SSL() bool {
	return mustParseBool(os.Getenv("S3_SECURE"))
}

// Convert string to bool and always return true if any error
func mustParseBool(str string) bool {
	b, err := strconv.ParseBool(str)
	if err != nil {
		return true
	}
	return b
}

func (c minioConfig) BucketName() string {
	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		// Default
		bucketName = "alice"
	}
	return bucketName
}

func (c minioConfig) Region() string {
	region := os.Getenv("S3_REGION")
	if region == "" {
		// Default
		region = "us-east-1"
	}
	return region
}

// PresignedPOST holds the remote URL and the policy data
// to successfully complete client upload request.
type presignedPOST struct {
	// Remote storage URL.
	URL string `json:",omitempty"`

	// Form policy data used..
	FormData map[string]string `json:",omitempty"`
}

// Generates new presigned POST policy data and the URL.
func (v *xrayHandlers) newPresignedURL(objPrefix string) (*presignedPOST, error) {
	// Set bucket and region obtained from configured minioConfig values..
	policy := minio.NewPostPolicy()
	policy.SetBucket(globalMinioClntConfig.BucketName())
	policy.SetRegion(globalMinioClntConfig.Region())

	// Set object key prefix where the images will be uploaded.
	policy.SetKeyStartsWith(objPrefix)

	// Expires in 10 days. (TODO make this configurable).
	policy.SetExpires(time.Now().UTC().AddDate(0, 0, 10))

	// Returns form data for POST form request.
	url, formData, err := v.minioClient.PresignedPostPolicy(policy)
	if err != nil {
		return nil, err
	}

	// Success..
	return &presignedPOST{
		URL:      url.String(),
		FormData: formData,
	}, nil
}

// Create a minio client to play.minio.io and make a bucket.
func newMinioClient() (*minio.Client, error) {
	// Initialize minio client instance.
	minioClient, err := minio.New(globalMinioClntConfig.Endpoint(), globalMinioClntConfig.AccessKey(),
		globalMinioClntConfig.SecretKey(), globalMinioClntConfig.SSL())
	if err != nil {
		return nil, err
	}

	// Check to see if we already own this bucket (which happens if you run this twice)
	exists, err := minioClient.BucketExists(globalMinioClntConfig.BucketName())
	if err != nil {
		return nil, err
	}

	// Create the bucket if it doesn't exist yet.
	if !exists {
		err = minioClient.MakeBucket(globalMinioClntConfig.BucketName(), globalMinioClntConfig.Region())
		if err != nil {
			return nil, err
		}
	}

	return minioClient, nil

}
