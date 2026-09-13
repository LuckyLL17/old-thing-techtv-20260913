package domain

import "time"

const (
	TutorialStatusDraft     = "draft"
	TutorialStatusPublished = "published"
	TutorialStatusArchived  = "archived"
	DifficultyEasy          = "easy"
	DifficultyMedium        = "medium"
	DifficultyHard          = "hard"
	CatFurniture  = "furniture"
	CatClothing   = "clothing"
	CatContainer  = "container"
	CatDecoration = "decoration"
	CatTool       = "tool"
	CatElectronic = "electronic"
	CatOther      = "other"
)

type Tutorial struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint64    `gorm:"index;not null" json:"user_id"`
	User          *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CategoryID    uint64    `gorm:"index;not null" json:"category_id"`
	Category      *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Title         string    `gorm:"size:200;not null" json:"title"`
	Slug          string    `gorm:"size:200;index" json:"slug"`
	Summary       string    `gorm:"type:text" json:"summary"`
	CoverBefore   string    `gorm:"size:255;not null" json:"cover_before"`
	CoverAfter    string    `gorm:"size:255;not null" json:"cover_after"`
	Difficulty    string    `gorm:"size:20;default:medium" json:"difficulty"`
	EstimatedHours float64  `gorm:"default:1" json:"estimated_hours"`
	Status        string    `gorm:"size:20;default:draft" json:"status"`
	Version       int       `gorm:"default:1" json:"version"`
	ViewCount     int       `gorm:"default:0;index" json:"view_count"`
	FavoriteCount int       `gorm:"default:0" json:"favorite_count"`
	AttemptCount  int       `gorm:"default:0" json:"attempt_count"`
	CommentCount  int       `gorm:"default:0" json:"comment_count"`
	ProjectCount  int       `gorm:"default:0" json:"project_count"`
	Tags          []*Tag    `gorm:"many2many:tutorial_tags;" json:"tags,omitempty"`
	Steps         []*Step   `gorm:"foreignKey:TutorialID" json:"steps,omitempty"`
	Materials     []*Material `gorm:"foreignKey:TutorialID" json:"materials,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

func (t *Tutorial) TableName() string {
	return "tutorials"
}

type TutorialVersion struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TutorialID  uint64    `gorm:"index" json:"tutorial_id"`
	Version     int       `json:"version"`
	Title       string    `gorm:"size:200" json:"title"`
	Summary     string    `gorm:"type:text" json:"summary"`
	ContentDump string    `gorm:"type:text" json:"content_dump"`
	CreatedAt   time.Time `json:"created_at"`
}

func (TutorialVersion) TableName() string {
	return "tutorial_versions"
}
