package repository

import (
	"strings"
	"upcycle-hub/internal/domain"
	apperr "upcycle-hub/pkg/errors"

	"gorm.io/gorm"
)

// isDuplicateKey 识别唯一索引冲突。
// SQLite(mattn): "UNIQUE constraint failed: ..."；MySQL: Error 1062
func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") || strings.Contains(msg, "Duplicate entry")
}

type BadgeRepo struct {
	db *gorm.DB
}

func NewBadgeRepo(db *gorm.DB) *BadgeRepo {
	return &BadgeRepo{db: db}
}

func (r *BadgeRepo) DB() *gorm.DB {
	return r.db
}

// Exists 判断用户是否已拥有某枚徽章（重复达成不发第二枚的第一道检查）
func (r *BadgeRepo) Exists(userID uint64, badgeCode string) (bool, error) {
	var n int64
	err := r.db.Model(&domain.UserBadge{}).
		Where("user_id = ? AND badge_code = ?", userID, badgeCode).
		Count(&n).Error
	if err != nil {
		return false, apperr.Wrap(apperr.CodeDB, "查询用户徽章失败", err)
	}
	return n > 0, nil
}

// Grant 颁发徽章。uk_user_badge 唯一索引兜底并发场景下的重复颁发：
// 命中唯一约束时返回 (false, nil)，调用方按重复处理。
func (r *BadgeRepo) Grant(ub *domain.UserBadge) (bool, error) {
	res := r.db.Create(ub)
	if res.Error != nil {
		if isDuplicateKey(res.Error) {
			return false, nil
		}
		return false, apperr.Wrap(apperr.CodeDB, "颁发徽章失败", res.Error)
	}
	return res.RowsAffected > 0, nil
}

// ListByUser 查询用户已获得的徽章，按徽章定义排序
func (r *BadgeRepo) ListByUser(userID uint64) ([]*domain.UserBadge, error) {
	var list []*domain.UserBadge
	err := r.db.Where("user_id = ?", userID).Order("id ASC").Find(&list).Error
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDB, "查询用户徽章失败", err)
	}
	return list, nil
}

// ListAllUserIDs 全量对账时遍历所有用户
func (r *BadgeRepo) ListAllUserIDs() ([]uint64, error) {
	var ids []uint64
	err := r.db.Model(&domain.User{}).Where("status = ?", 1).Pluck("id", &ids).Error
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDB, "查询用户列表失败", err)
	}
	return ids, nil
}

// CreateLog 写入一条判定/发放流水
func (r *BadgeRepo) CreateLog(l *domain.BadgeAwardLog) error {
	if err := r.db.Create(l).Error; err != nil {
		return apperr.Wrap(apperr.CodeDB, "写入徽章发放流水失败", err)
	}
	return nil
}

// ListLogs 查询发放流水（管理追踪用）
func (r *BadgeRepo) ListLogs(page, size int, userID uint64, badgeCode, result, event string) ([]*domain.BadgeAwardLog, int64, error) {
	var list []*domain.BadgeAwardLog
	var total int64
	q := r.db.Model(&domain.BadgeAwardLog{})
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if badgeCode != "" {
		q = q.Where("badge_code = ?", badgeCode)
	}
	if result != "" {
		q = q.Where("result = ?", result)
	}
	if event != "" {
		q = q.Where("event = ?", event)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeDB, "统计徽章流水失败", err)
	}
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 30
	}
	err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeDB, "查询徽章流水失败", err)
	}
	return list, total, nil
}

// ---- 成就指标采集 ----

// CountPublishedTutorials 用户已发布教程数
func (r *BadgeRepo) CountPublishedTutorials(userID uint64) (int, error) {
	var n int64
	err := r.db.Model(&domain.Tutorial{}).
		Where("user_id = ? AND status = ? AND deleted_at IS NULL", userID, domain.TutorialStatusPublished).
		Count(&n).Error
	return int(n), err
}

// MaxProjectLikes 用户单个作品的最高点赞数及对应作品 ID
func (r *BadgeRepo) MaxProjectLikes(userID uint64) (int, uint64, error) {
	var ps []domain.Project
	err := r.db.Where("user_id = ? AND status = ?", userID, 1).
		Order("like_count DESC, id ASC").Limit(1).Find(&ps).Error
	if err != nil {
		return 0, 0, err
	}
	if len(ps) == 0 {
		return 0, 0, nil
	}
	return ps[0].LikeCount, ps[0].ID, nil
}

// PublishedMonths 用户有教程发布的月份列表（YYYY-mm，升序去重）
func (r *BadgeRepo) PublishedMonths(userID uint64) ([]string, error) {
	var months []string
	err := r.db.Model(&domain.Tutorial{}).
		Where("user_id = ? AND status = ? AND deleted_at IS NULL", userID, domain.TutorialStatusPublished).
		Distinct("strftime('%Y-%m', created_at)").
		Order("1 ASC").Pluck("strftime('%Y-%m', created_at)", &months).Error
	if err != nil {
		return nil, err
	}
	return months, nil
}

// CountFavoriteOthers 收藏他人教程的数量（排除收藏自己的教程，去重）
func (r *BadgeRepo) CountFavoriteOthers(userID uint64) (int, error) {
	var n int64
	// 用子查询而非 JOIN：避免 GORM 软删除作用域给 JOIN 子句附加条件，
	// 也避免 favorites 与 tutorials 行乘积带来的计数歧义。
	err := r.db.Model(&domain.Favorite{}).
		Where("favorites.user_id = ? AND favorites.target_type = ?", userID, domain.FavTypeTutorial).
		Where("target_id IN (?)",
			r.db.Model(&domain.Tutorial{}).Select("id").
				Where("user_id <> ? AND status = ?", userID, domain.TutorialStatusPublished)).
		Count(&n).Error
	if err != nil {
		return 0, err
	}
	return int(n), nil
}
