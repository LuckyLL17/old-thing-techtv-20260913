package service

import (
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
)

type StatsService struct {
	tutorialRepo *repository.TutorialRepo
	projectRepo  *repository.ProjectRepo
	userRepo     *repository.UserRepo
	categoryRepo *repository.CategoryRepo
	favoriteRepo *repository.FavoriteRepo
	attemptRepo  *repository.AttemptRepo
}

func NewStatsService(tr *repository.TutorialRepo, pr *repository.ProjectRepo, ur *repository.UserRepo, cr *repository.CategoryRepo, fr *repository.FavoriteRepo, ar *repository.AttemptRepo) *StatsService {
	if fr == nil {
		fr = repository.NewFavoriteRepo(ur.DB())
	}
	if ar == nil {
		ar = repository.NewAttemptRepo(ur.DB())
	}
	return &StatsService{tutorialRepo: tr, projectRepo: pr, userRepo: ur, categoryRepo: cr, favoriteRepo: fr, attemptRepo: ar}
}

type DashboardStats struct {
	TutorialCount   int64              `json:"tutorial_count"`
	ProjectCount    int64              `json:"project_count"`
	UserCount       int64              `json:"user_count"`
	TopTutorials    []*domain.Tutorial `json:"top_tutorials"`
	TopUsers        []*domain.User     `json:"top_users"`
	CategoryStats   map[string]int64   `json:"category_stats"`
	MonthlyTrend    []int64            `json:"monthly_trend"`
	PublishedCount  int64              `json:"published_count"`
}

func (s *StatsService) Dashboard() (*DashboardStats, error) {
	st := &DashboardStats{}
	var err error
	st.TutorialCount, err = s.tutorialRepo.Count("")
	if err != nil {
		return nil, err
	}
	st.PublishedCount, err = s.tutorialRepo.Count(domain.TutorialStatusPublished)
	if err != nil {
		return nil, err
	}
	st.ProjectCount, err = s.projectRepo.Count()
	if err != nil {
		return nil, err
	}
	st.UserCount, err = s.userRepo.Count()
	if err != nil {
		return nil, err
	}
	st.TopTutorials, err = s.tutorialRepo.TopTutorials(10)
	if err != nil {
		return nil, err
	}
	st.TopUsers, err = s.userRepo.TopUsers(10)
	if err != nil {
		return nil, err
	}
	byCat, err := s.tutorialRepo.CountByCategory()
	if err == nil {
		cats, cerr := s.categoryRepo.ListAll()
		if cerr == nil {
			st.CategoryStats = make(map[string]int64)
			for _, c := range cats {
				st.CategoryStats[c.Name] = byCat[c.ID]
			}
		}
	}
	trend, err := s.tutorialRepo.CountByMonth(6)
	if err == nil {
		st.MonthlyTrend = trend
	}
	return st, nil
}

type UserCenterStats struct {
	TutorialCount  int    `json:"tutorial_count"`
	ProjectCount   int    `json:"project_count"`
	FavoriteCount  int64  `json:"favorite_count"`
	AttemptCount   int64  `json:"attempt_count"`
	CompletedCount int64  `json:"completed_count"`
	TotalItems     int    `json:"total_items"`
	Level          string `json:"level"`
	Score          int    `json:"score"`
}

func (s *StatsService) UserCenter(userID uint64) (*UserCenterStats, error) {
	u, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	st := &UserCenterStats{
		TutorialCount: u.TutorialCount,
		ProjectCount:  u.ProjectCount,
		Level:         u.Level,
		Score:         u.Score,
		TotalItems:    u.TutorialCount + u.ProjectCount,
	}
	fm, _ := s.favoriteRepo.CountByUser(userID)
	st.FavoriteCount = fm[domain.FavTypeTutorial] + fm[domain.FavTypeProject]
	st.AttemptCount, st.CompletedCount, _ = s.attemptRepo.CountByUser(userID)
	return st, nil
}
