//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import (
	"time"
)

type ImageUploads struct {
	ID         int64 `sql:"primary_key"`
	UploadName *string
	UserID     int64
	Imagepath  string
	UploadedAt time.Time
}
