package datanode

import (
	"time"

	"github.com/jinzhu/gorm"
)

// File Represents the file info model
type File struct {
	gorm.Model
	Token          string     `gorm:"unique_index;not null"` //Unique token for the file
	Name           string     //Name of file
	Type           string     //ndicates type of file (video, model .... etc)
	Path           string     //Path to file (excluding file name)
	HLSPath        string     //Path to HLS file in case of video
	ThumbnailPath  string     //Path to thumbnail of video (in case of video)
	Extras         string     `gorm:"size:500"` //Extras json field for any extra info
	DataNodeID     string     //ID of the data node that has the file
	Parent         string     //Token of the parent file in case it's a replica
	Offset         int64      //Offset of bytes to start writing data at
	Size           int64      //Total size of file in bytes
	TotalJobCount  int        //Total number of jobs needed to apply on video
	TotalDoneCount int        //Number of jobs applied to the video
	CompletedAt    *time.Time //Indicates if file completed uploading
}
