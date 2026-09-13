package domain

import "time"

const (
	UserLevelNovice = "novice"
	UserLevelApprentice = "apprentice"
	UserLevelCraftsman  = "craftsman"
	UserLevelMaster     = "master"
)

type User struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username      string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Email         string    `gorm:"size:100;uniqueIndex;not null" json:"email"`
	PasswordHash  string    `gorm:"size:255;not null" json:"-"`
	Avatar        string    `gorm:"size:255" json:"avatar"`
	Nickname      string    `gorm:"size:50" json:"nickname"`
	Specialty     string    `gorm:"size:255" json:"specialty"`
	Bio           string    `gorm:"type:text" json:"bio"`
	Level         string    `gorm:"size:20;default:novice" json:"level"`
	TutorialCount int       `gorm:"default:0" json:"tutorial_count"`
	ProjectCount  int       `gorm:"default:0" json:"project_count"`
	Score         int       `gorm:"default:0" json:"score"`
	Status        int       `gorm:"default:1" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) ComputeLevel() {
	if u.Score >= 5000 || u.TutorialCount >= 50 {
		u.Level = UserLevelMaster
	} else if u.Score >= 1000 || u.TutorialCount >= 15 {
		u.Level = UserLevelCraftsman
	} else if u.Score >= 200 || u.TutorialCount >= 3 {
		u.Level = UserLevelApprentice
	} else {
		u.Level = UserLevelNovice
	}
}

func (u *User) DisplayName() string {
	if u.Nickname != "" {
		return u.Nickname
	}
	return u.Username
}
