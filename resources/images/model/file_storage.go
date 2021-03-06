// Copyright 2017 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package model

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/mendersoftware/deployments/resources/images"
)

// Errors specific to interface
var (
	ErrFileStorageFileNotFound = errors.New("File not found")
)

// FileStorage allows to store and manage large files
type FileStorage interface {
	Delete(ctx context.Context, objectId string) error
	Exists(ctx context.Context, objectId string) (bool, error)
	LastModified(ctx context.Context, objectId string) (time.Time, error)
	PutRequest(ctx context.Context, objectId string,
		duration time.Duration) (*images.Link, error)
	GetRequest(ctx context.Context, objectId string,
		duration time.Duration, responseContentType string) (*images.Link, error)
	UploadArtifact(ctx context.Context, objectId string,
		artifactSize int64, artifact io.Reader, contentType string) error
}
