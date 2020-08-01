package namenode

import (
	"github.com/jinzhu/gorm"
)

// Clip Represents a clip record in the ingestion database
type Clip struct {
	gorm.Model
	Token     string //The token of the video to which this clip belongs
	Tag       string `gorm:"index:idx_clips_tag_start_end"` //The classification of the clip
	StartTime uint64 `gorm:"index:idx_clips_tag_start_end"` //Start time of the clip's tag (secs from start)
	EndTime   uint64 `gorm:"index:idx_clips_tag_start_end"` //End time for the clip's tag (secs from start)
}
