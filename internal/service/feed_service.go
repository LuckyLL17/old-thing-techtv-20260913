package service

import (
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
)

// FeedItem 关注动态中的一条混排条目（教程或改造作品）
type FeedItem struct {
	Type      string           `json:"type"` // tutorial | project
	Tutorial  *domain.Tutorial `json:"tutorial,omitempty"`
	Project   *domain.Project  `json:"project,omitempty"`
	CreatedAt string           `json:"created_at"`
}

// FeedResult 一页关注动态
type FeedResult struct {
	Items          []*FeedItem
	Total          int64
	Page, Size     int
	HasMore        bool
	FollowingCount int
	Suggestions    []*domain.User
}

type FeedService struct {
	feedRepo     *repository.FeedRepo
	followRepo   *repository.FollowRepo
	tutorialRepo *repository.TutorialRepo
	projectRepo  *repository.ProjectRepo
	userRepo     *repository.UserRepo
}

func NewFeedService(fr *repository.FeedRepo, flr *repository.FollowRepo,
	tur *repository.TutorialRepo, pr *repository.ProjectRepo, ur *repository.UserRepo) *FeedService {
	return &FeedService{feedRepo: fr, followRepo: flr, tutorialRepo: tur, projectRepo: pr, userRepo: ur}
}

func (s *FeedService) Feed(userID uint64, page, size int) (*FeedResult, error) {
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 50 {
		size = 10
	}
	following, err := s.followRepo.Following(userID)
	if err != nil {
		return nil, err
	}
	res := &FeedResult{Page: page, Size: size, FollowingCount: len(following), Items: []*FeedItem{}}

	// 没有关注任何作者：返回发现作者引导所需的推荐列表
	if len(following) == 0 {
		res.Suggestions, err = s.userRepo.SuggestAuthors([]uint64{userID}, 8)
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	refs, total, err := s.feedRepo.FeedRefs(following, page, size)
	if err != nil {
		return nil, err
	}
	res.Total = total
	res.HasMore = int64(page*size) < total

	var tutIDs, projIDs []uint64
	for _, ref := range refs {
		switch ref.ItemType {
		case "tutorial":
			tutIDs = append(tutIDs, ref.ID)
		case "project":
			projIDs = append(projIDs, ref.ID)
		}
	}
	tutMap := make(map[uint64]*domain.Tutorial)
	if len(tutIDs) > 0 {
		tuts, err := s.tutorialRepo.ListPublishedByIDs(tutIDs)
		if err != nil {
			return nil, err
		}
		for _, t := range tuts {
			tutMap[t.ID] = t
		}
	}
	projMap := make(map[uint64]*domain.Project)
	if len(projIDs) > 0 {
		projs, err := s.projectRepo.ListActiveByIDs(projIDs)
		if err != nil {
			return nil, err
		}
		for _, p := range projs {
			projMap[p.ID] = p
		}
	}

	// 严格按 refs 的时间倒序拼装；取不到（被删/状态变更）则跳过，绝不出现非公开内容
	for _, ref := range refs {
		switch ref.ItemType {
		case "tutorial":
			if t := tutMap[ref.ID]; t != nil {
				res.Items = append(res.Items, &FeedItem{Type: "tutorial", Tutorial: t, CreatedAt: ref.CreatedAt.Format("2006-01-02T15:04:05Z07:00")})
			}
		case "project":
			if p := projMap[ref.ID]; p != nil {
				res.Items = append(res.Items, &FeedItem{Type: "project", Project: p, CreatedAt: ref.CreatedAt.Format("2006-01-02T15:04:05Z07:00")})
			}
		}
	}
	return res, nil
}
