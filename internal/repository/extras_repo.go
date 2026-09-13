package repository

import (
	"time"
	"upcycle-hub/internal/domain"
	apperr "upcycle-hub/pkg/errors"

	"gorm.io/gorm"
)

type TutorialVersionRepo struct {
	db *gorm.DB
}

func NewTutorialVersionRepo(db *gorm.DB) *TutorialVersionRepo {
	return &TutorialVersionRepo{db: db}
}

func (r *TutorialVersionRepo) Create(v *domain.TutorialVersion) error {
	err := r.db.Create(v).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "创建教程版本失败", err)
	}
	return nil
}

func (r *TutorialVersionRepo) GetByID(id uint64) (*domain.TutorialVersion, error) {
	v := &domain.TutorialVersion{}
	err := r.db.First(v, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询教程版本失败", err)
	}
	return v, nil
}

func (r *TutorialVersionRepo) ListByTutorial(tid uint64, page, size int) ([]*domain.TutorialVersion, int64, error) {
	var list []*domain.TutorialVersion
	var total int64
	q := r.db.Model(&domain.TutorialVersion{}).Where("tutorial_id = ?", tid)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeDB, "统计版本数量失败", err)
	}
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size
	err := q.Order("version DESC, id DESC").Offset(offset).Limit(size).Find(&list).Error
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeDB, "查询教程版本列表失败", err)
	}
	return list, total, nil
}

func (r *TutorialVersionRepo) GetLatest(tid uint64) (*domain.TutorialVersion, error) {
	v := &domain.TutorialVersion{}
	err := r.db.Where("tutorial_id = ?", tid).Order("version DESC, id DESC").First(v).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询最新版本失败", err)
	}
	return v, nil
}

func (r *TutorialVersionRepo) GetByVersion(tid uint64, version int) (*domain.TutorialVersion, error) {
	v := &domain.TutorialVersion{}
	err := r.db.Where("tutorial_id = ? AND version = ?", tid, version).First(v).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询指定版本失败", err)
	}
	return v, nil
}

type NotificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepo(db *gorm.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func (r *NotificationRepo) Create(n *domain.Notification) error {
	err := r.db.Create(n).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "创建通知失败", err)
	}
	return nil
}

func (r *NotificationRepo) BatchCreate(list []*domain.Notification) error {
	if len(list) == 0 {
		return nil
	}
	err := r.db.Create(&list).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "批量创建通知失败", err)
	}
	return nil
}

func (r *NotificationRepo) ListByUser(uid uint64, page, size int, onlyUnread bool) ([]*domain.Notification, int64, error) {
	var list []*domain.Notification
	var total int64
	q := r.db.Model(&domain.Notification{}).Where("user_id = ?", uid)
	if onlyUnread {
		q = q.Where("`read` = ?", false)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeDB, "统计通知数量失败", err)
	}
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	offset := (page - 1) * size
	err := q.Preload("Actor").Order("id DESC").Offset(offset).Limit(size).Find(&list).Error
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeDB, "查询通知列表失败", err)
	}
	return list, total, nil
}

func (r *NotificationRepo) CountUnread(uid uint64) (int64, error) {
	var total int64
	err := r.db.Model(&domain.Notification{}).Where("user_id = ? AND `read` = ?", uid, false).Count(&total).Error
	if err != nil {
		return 0, apperr.Wrap(apperr.CodeDB, "统计未读通知失败", err)
	}
	return total, nil
}

func (r *NotificationRepo) MarkRead(uid, id uint64) error {
	now := time.Now()
	res := r.db.Model(&domain.Notification{}).
		Where("id = ? AND user_id = ?", id, uid).
		Updates(map[string]any{"`read`": true, "read_at": &now})
	if res.Error != nil {
		return apperr.Wrap(apperr.CodeDB, "标记通知已读失败", res.Error)
	}
	return nil
}

func (r *NotificationRepo) MarkAllRead(uid uint64) (int64, error) {
	now := time.Now()
	res := r.db.Model(&domain.Notification{}).
		Where("user_id = ? AND `read` = ?", uid, false).
		Updates(map[string]any{"`read`": true, "read_at": &now})
	if res.Error != nil {
		return 0, apperr.Wrap(apperr.CodeDB, "标记全部已读失败", res.Error)
	}
	return res.RowsAffected, nil
}

func (r *NotificationRepo) DeleteByUser(uid uint64, before *time.Time) (int64, error) {
	q := r.db.Where("user_id = ?", uid)
	if before != nil {
		q = q.Where("created_at < ?", *before)
	}
	res := q.Delete(&domain.Notification{})
	if res.Error != nil {
		return 0, apperr.Wrap(apperr.CodeDB, "清理通知失败", res.Error)
	}
	return res.RowsAffected, nil
}

type AuditLogRepo struct {
	db *gorm.DB
}

func NewAuditLogRepo(db *gorm.DB) *AuditLogRepo {
	return &AuditLogRepo{db: db}
}

func (r *AuditLogRepo) Create(l *domain.AuditLog) error {
	err := r.db.Create(l).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "创建审计日志失败", err)
	}
	return nil
}

func (r *AuditLogRepo) List(page, size int, userID uint64, action, targetType string, from, to *time.Time) ([]*domain.AuditLog, int64, error) {
	var list []*domain.AuditLog
	var total int64
	q := r.db.Model(&domain.AuditLog{})
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if targetType != "" {
		q = q.Where("target_type = ?", targetType)
	}
	if from != nil {
		q = q.Where("created_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("created_at <= ?", *to)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeDB, "统计审计日志数量失败", err)
	}
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 500 {
		size = 30
	}
	offset := (page - 1) * size
	err := q.Preload("User").Order("id DESC").Offset(offset).Limit(size).Find(&list).Error
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeDB, "查询审计日志失败", err)
	}
	return list, total, nil
}

func (r *AuditLogRepo) StatsByAction(days int) (map[string]int64, error) {
	result := map[string]int64{}
	type row struct {
		Action string `gorm:"column:action"`
		Count  int64  `gorm:"column:cnt"`
	}
	var rows []row
	q := r.db.Model(&domain.AuditLog{})
	if days > 0 {
		since := time.Now().AddDate(0, 0, -days)
		q = q.Where("created_at >= ?", since)
	}
	err := q.Select("action, COUNT(*) AS cnt").Group("action").Scan(&rows).Error
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDB, "统计审计动作失败", err)
	}
	for _, r := range rows {
		result[r.Action] = r.Count
	}
	return result, nil
}
