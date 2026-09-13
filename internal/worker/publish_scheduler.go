package worker

import (
	"context"
	"time"
	"upcycle-hub/internal/service"
	"upcycle-hub/pkg/logger"
)

// PublishScheduler 周期扫描到期的定时教程并自动发布。
// 启动时立即执行一次，确保服务重启期间到期的任务不会漏发。
type PublishScheduler struct {
	tutorialSvc *service.TutorialService
	interval    time.Duration
	stop        chan struct{}
}

func NewPublishScheduler(svc *service.TutorialService) *PublishScheduler {
	return &PublishScheduler{tutorialSvc: svc, interval: 30 * time.Second, stop: make(chan struct{})}
}

func (w *PublishScheduler) Start(ctx context.Context) {
	go w.loop(ctx)
	logger.Infof("publish scheduler started")
}

func (w *PublishScheduler) Stop() {
	close(w.stop)
}

func (w *PublishScheduler) loop(ctx context.Context) {
	t := time.NewTicker(w.interval)
	defer t.Stop()
	w.tutorialSvc.PublishDue()
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stop:
			return
		case <-t.C:
			w.tutorialSvc.PublishDue()
		}
	}
}
