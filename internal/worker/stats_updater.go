package worker

import (
	"context"
	"time"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
	"upcycle-hub/pkg/logger"
)

type StatsUpdater struct {
	userRepo     *repository.UserRepo
	tutorialRepo *repository.TutorialRepo
	commentRepo  *repository.CommentRepo
	catRepo      *repository.CategoryRepo
	tagRepo      *repository.TagRepo
	stop         chan struct{}
}

func NewStatsUpdater(ur *repository.UserRepo, tr *repository.TutorialRepo,
	cr *repository.CommentRepo, car *repository.CategoryRepo, tgr *repository.TagRepo) *StatsUpdater {
	return &StatsUpdater{userRepo: ur, tutorialRepo: tr, commentRepo: cr, catRepo: car, tagRepo: tgr, stop: make(chan struct{})}
}

func (w *StatsUpdater) Start(ctx context.Context) {
	go w.loop(ctx)
	logger.Infof("stats updater started")
}

func (w *StatsUpdater) Stop() {
	close(w.stop)
}

func (w *StatsUpdater) loop(ctx context.Context) {
	t := time.NewTicker(10 * time.Minute)
	defer t.Stop()
	w.RunOnce()
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stop:
			return
		case <-t.C:
			w.RunOnce()
		}
	}
}

func (w *StatsUpdater) RunOnce() {
	w.refreshUserLevels()
	w.syncCategoryCounts()
	w.syncTagCounts()
	logger.Infof("stats update run complete")
}

func (w *StatsUpdater) refreshUserLevels() {
	page := 1
	size := 100
	for {
		list, _, err := w.userRepo.List(page, size, "")
		if err != nil {
			logger.Errorf("refreshUserLevels list: %v", err)
			return
		}
		if len(list) == 0 {
			break
		}
		for _, u := range list {
			old := u.Level
			u.ComputeLevel()
			if u.Level != old {
				if err := w.userRepo.Update(u); err != nil {
					logger.Errorf("update user %d level: %v", u.ID, err)
				}
			}
		}
		page++
		if len(list) < size {
			break
		}
	}
}

func (w *StatsUpdater) syncCategoryCounts() {
	counts, err := w.tutorialRepo.CountByCategory()
	if err != nil {
		logger.Errorf("CountByCategory: %v", err)
		return
	}
	cats, err := w.catRepo.ListAll()
	if err != nil {
		return
	}
	for _, c := range cats {
		real := counts[c.ID]
		if int64(c.TutorialCount) != real {
			c.TutorialCount = int(real)
			w.catRepo.Update(c)
		}
	}
}

func (w *StatsUpdater) syncTagCounts() {
	tags, err := w.tagRepo.ListAll(0)
	if err != nil {
		return
	}
	for _, t := range tags {
		var n int64
		w.tutorialRepo.DB().Model(&domain.TutorialTag{}).Where("tag_id = ?", t.ID).Count(&n)
		if int64(t.TutorialCount) != n {
			w.tagRepo.IncCount(t.ID, int(n)-t.TutorialCount)
		}
	}
}
