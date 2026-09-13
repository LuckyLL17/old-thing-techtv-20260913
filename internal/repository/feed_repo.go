package repository

import (
	"fmt"
	"strings"
	"time"
	"upcycle-hub/internal/domain"
	apperr "upcycle-hub/pkg/errors"

	"gorm.io/gorm"
)

// FeedRef 是跨表时间线的轻量指针：条目类型 + 主键 + 发布时间
type FeedRef struct {
	ItemType  string    `gorm:"column:item_type"`
	ID        uint64    `gorm:"column:id"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

type FeedRepo struct {
	db *gorm.DB
}

func NewFeedRepo(db *gorm.DB) *FeedRepo {
	return &FeedRepo{db: db}
}

// FeedRefs 把所关注作者的已发布教程与正常状态作品做 UNION ALL 混排，
// 按发布时间倒序分页，返回本页指针与总数
func (r *FeedRepo) FeedRefs(followingIDs []uint64, page, size int) ([]FeedRef, int64, error) {
	if len(followingIDs) == 0 {
		return []FeedRef{}, 0, nil
	}
	ph := "?" + strings.Repeat(",?", len(followingIDs)-1)
	idArgs := make([]interface{}, len(followingIDs))
	for i, id := range followingIDs {
		idArgs[i] = id
	}

	countSQL := `
SELECT COUNT(*) FROM (
  SELECT id FROM tutorials
   WHERE status = ? AND deleted_at IS NULL AND user_id IN (` + ph + `)
  UNION ALL
  SELECT id FROM projects
   WHERE status = 1 AND user_id IN (` + ph + `)
)`
	countArgs := append([]interface{}{domain.TutorialStatusPublished}, idArgs...)
	countArgs = append(countArgs, idArgs...)
	var total int64
	if err := r.db.Raw(countSQL, countArgs...).Scan(&total).Error; err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeDB, "统计关注动态失败", err)
	}
	if total == 0 || int64((page-1)*size) >= total {
		return []FeedRef{}, total, nil
	}

	// page/size 已在 service 层归一化为整数，直接内联进 SQL
	pageSQL := fmt.Sprintf(`
SELECT item_type, id, created_at FROM (
  SELECT 'tutorial' AS item_type, id, created_at FROM tutorials
   WHERE status = ? AND deleted_at IS NULL AND user_id IN (%s)
  UNION ALL
  SELECT 'project' AS item_type, id, created_at FROM projects
   WHERE status = 1 AND user_id IN (%s)
)
ORDER BY created_at DESC, item_type DESC, id DESC
LIMIT %d OFFSET %d`, ph, ph, size, (page-1)*size)
	args := append([]interface{}{domain.TutorialStatusPublished}, idArgs...)
	args = append(args, idArgs...)
	var refs []FeedRef
	if err := r.db.Raw(pageSQL, args...).Scan(&refs).Error; err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeDB, "查询关注动态失败", err)
	}
	return refs, total, nil
}
