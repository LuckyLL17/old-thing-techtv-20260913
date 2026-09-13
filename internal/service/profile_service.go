package service

import (
	"time"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
)

// profileListLimit 用户主页内联展示的教程/作品条数，更多走列表页分页
const profileListLimit = 30

// PublicUser 对外公开的用户信息，绝不包含邮箱、密码等敏感字段
type PublicUser struct {
	ID            uint64    `json:"id"`
	Username      string    `json:"username"`
	Avatar        string    `json:"avatar"`
	Nickname      string    `json:"nickname"`
	Specialty     string    `json:"specialty"`
	Bio           string    `json:"bio"`
	Level         string    `json:"level"`
	TutorialCount int       `json:"tutorial_count"`
	ProjectCount  int       `json:"project_count"`
	Score         int       `json:"score"`
	CreatedAt     time.Time `json:"created_at"`
}

func toPublicUser(u *domain.User) *PublicUser {
	return &PublicUser{
		ID:            u.ID,
		Username:      u.Username,
		Avatar:        u.Avatar,
		Nickname:      u.Nickname,
		Specialty:     u.Specialty,
		Bio:           u.Bio,
		Level:         u.Level,
		TutorialCount: u.TutorialCount,
		ProjectCount:  u.ProjectCount,
		Score:         u.Score,
		CreatedAt:     u.CreatedAt,
	}
}

// stripEmbeddedEmails 公开主页内联的教程/作品带有嵌套作者，清掉其邮箱避免泄漏
func stripEmbeddedEmails(tutorials []*domain.Tutorial, projects []*domain.Project) {
	for _, t := range tutorials {
		if t.User != nil {
			t.User.Email = ""
		}
	}
	for _, p := range projects {
		if p.User != nil {
			p.User.Email = ""
		}
	}
}

// UserProfile 用户主页聚合数据
type UserProfile struct {
	User          *PublicUser        `json:"user"`
	Followers     int64              `json:"followers"`
	Following     int64              `json:"following"`
	IsFollowing   bool               `json:"is_following"`
	IsSelf        bool               `json:"is_self"`
	Tutorials     []*domain.Tutorial `json:"tutorials"`
	Projects      []*domain.Project  `json:"projects"`
	TutorialTotal int64              `json:"tutorial_total"`
	ProjectTotal  int64              `json:"project_total"`
}

type ProfileService struct {
	userRepo     *repository.UserRepo
	tutorialRepo *repository.TutorialRepo
	projectRepo  *repository.ProjectRepo
	followRepo   *repository.FollowRepo
}

func NewProfileService(ur *repository.UserRepo, tur *repository.TutorialRepo, pr *repository.ProjectRepo, flr *repository.FollowRepo) *ProfileService {
	return &ProfileService{userRepo: ur, tutorialRepo: tur, projectRepo: pr, followRepo: flr}
}

// GetProfile 返回用户主页数据。viewerID 为 0 表示游客访问。
// 用户不存在返回 apperr.ErrUserNotFound；账号被禁用返回 403 业务错误。
func (s *ProfileService) GetProfile(viewerID, targetID uint64) (*UserProfile, error) {
	u, err := s.userRepo.GetByID(targetID)
	if err != nil {
		return nil, err
	}
	if u.Status != domain.UserStatusActive {
		return nil, ErrForbidden("该账号已被禁用，暂时无法查看其主页")
	}

	tutorials, tutorialTotal, err := s.tutorialRepo.List(1, profileListLimit, 0, "",
		domain.TutorialStatusPublished, "new", "", targetID)
	if err != nil {
		return nil, err
	}
	projects, projectTotal, err := s.projectRepo.List(1, profileListLimit, 0, targetID, "")
	if err != nil {
		return nil, err
	}
	followers, following, err := s.followRepo.Counts(targetID)
	if err != nil {
		return nil, err
	}
	isFollowing := false
	if viewerID > 0 && viewerID != targetID {
		isFollowing, _ = s.followRepo.IsFollowing(viewerID, targetID)
	}
	if tutorials == nil {
		tutorials = []*domain.Tutorial{}
	}
	if projects == nil {
		projects = []*domain.Project{}
	}
	stripEmbeddedEmails(tutorials, projects)

	return &UserProfile{
		User:          toPublicUser(u),
		Followers:     followers,
		Following:     following,
		IsFollowing:   isFollowing,
		IsSelf:        viewerID == targetID,
		Tutorials:     tutorials,
		Projects:      projects,
		TutorialTotal: tutorialTotal,
		ProjectTotal:  projectTotal,
	}, nil
}
