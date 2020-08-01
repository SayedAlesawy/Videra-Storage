package namenode

import (
	"github.com/jinzhu/gorm"
)

// Clip Represents a clip record in the ingestion database
type Clip struct {
	gorm.Model
	Token     string `json:"token"`                                           //The token of the video to which this clip belongs
	Tag       string `gorm:"index:idx_clips_tag_start_end" json:"tag"`        //The classification of the clip
	StartTime uint64 `gorm:"index:idx_clips_tag_start_end" json:"start_time"` //Start time of the clip's tag (secs from start)
	EndTime   uint64 `gorm:"index:idx_clips_tag_start_end" json:"end_time"`   //End time for the clip's tag (secs from start)
}
