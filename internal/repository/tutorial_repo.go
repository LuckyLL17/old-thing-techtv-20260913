package repository

import (
	"upcycle-hub/internal/domain"
	apperr "upcycle-hub/pkg/errors"

	"gorm.io/gorm"
)

type TutorialRepo struct {
	db *gorm.DB
}

func NewTutorialRepo(db *gorm.DB) *TutorialRepo {
	return &TutorialRepo{db: db}
}

func (r *TutorialRepo) DB() *gorm.DB {
	return r.db
}

func (r *TutorialRepo) Create(t *domain.Tutorial) error {
	err := r.db.Create(t).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "创建教程失败", err)
	}
	return nil
}

func (r *TutorialRepo) GetByID(id uint64, full bool) (*domain.Tutorial, error) {
	t := &domain.Tutorial{}
	q := r.db.Preload("User").Preload("Category")
	if full {
		q = q.Preload("Tags").Preload("Steps", func(d *gorm.DB) *gorm.DB {
			return d.Order("step_order ASC")
		}).Preload("Materials", func(d *gorm.DB) *gorm.DB {
			return d.Order("sort_order ASC")
		})
	}
	err := q.First(t, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.ErrTutorialNotFound
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询教程失败", err)
	}
	return t, nil
}

func (r *TutorialRepo) Update(t *domain.Tutorial) error {
	err := r.db.Save(t).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "更新教程失败", err)
	}
	return nil
}

func (r *TutorialRepo) Delete(id uint64) error {
	err := r.db.Delete(&domain.Tutorial{}, id).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "删除教程失败", err)
	}
	return nil
}

func (r *TutorialRepo) List(page, size int, catID uint64, diff, status, sort, keyword string, userID uint64) ([]*domain.Tutorial, int64, error) {
	var total int64
	var list []*domain.Tutorial
	q := r.db.Model(&domain.Tutorial{}).Preload("User").Preload("Category").Preload("Tags")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if catID > 0 {
		q = q.Where("category_id = ?", catID)
	}
	if diff != "" {
		q = q.Where("difficulty = ?", diff)
	}
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if keyword != "" {
		q = q.Where("title LIKE ? OR summary LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	q.Count(&total)
	orderExpr := "id DESC"
	switch sort {
	case "popular":
		orderExpr = "favorite_count DESC, view_count DESC"
	case "attempted":
		orderExpr = "attempt_count DESC"
	case "views":
		orderExpr = "view_count DESC"
	case "new":
		orderExpr = "created_at DESC"
	}
	if size > 0 {
		q = q.Offset((page - 1) * size).Limit(size)
	}
	err := q.Order(orderExpr).Find(&list).Error
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeDB, "查询教程列表失败", err)
	}
	return list, total, nil
}

func (r *TutorialRepo) IncView(id uint64) error {
	err := r.db.Model(&domain.Tutorial{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "增加浏览数失败", err)
	}
	return nil
}

func (r *TutorialRepo) IncCounts(id uint64, fav, attempt, comment, project int) error {
	updates := map[string]interface{}{}
	if fav != 0 {
		updates["favorite_count"] = gorm.Expr("favorite_count + ?", fav)
	}
	if attempt != 0 {
		updates["attempt_count"] = gorm.Expr("attempt_count + ?", attempt)
	}
	if comment != 0 {
		updates["comment_count"] = gorm.Expr("comment_count + ?", comment)
	}
	if project != 0 {
		updates["project_count"] = gorm.Expr("project_count + ?", project)
	}
	if len(updates) == 0 {
		return nil
	}
	err := r.db.Model(&domain.Tutorial{}).Where("id = ?", id).Updates(updates).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "更新统计失败", err)
	}
	return nil
}

func (r *TutorialRepo) TopTutorials(n int) ([]*domain.Tutorial, error) {
	var list []*domain.Tutorial
	err := r.db.Preload("User").Preload("Category").Where("status = ?", domain.TutorialStatusPublished).
		Order("favorite_count DESC, view_count DESC").Limit(n).Find(&list).Error
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDB, "查询热门教程失败", err)
	}
	return list, nil
}

func (r *TutorialRepo) Random(n int) ([]*domain.Tutorial, error) {
	var list []*domain.Tutorial
	err := r.db.Preload("User").Preload("Category").
		Where("status = ?", domain.TutorialStatusPublished).
		Order("RANDOM()").Limit(n).Find(&list).Error
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDB, "随机推荐失败", err)
	}
	return list, nil
}

func (r *TutorialRepo) Count(status string) (int64, error) {
	var n int64
	q := r.db.Model(&domain.Tutorial{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Count(&n).Error
	return n, err
}

func (r *TutorialRepo) CountByCategory() (map[uint64]int64, error) {
	rows, err := r.db.Model(&domain.Tutorial{}).Select("category_id, count(*)").
		Where("status = ?", domain.TutorialStatusPublished).
		Group("category_id").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[uint64]int64)
	for rows.Next() {
		var cid uint64
		var n int64
		rows.Scan(&cid, &n)
		m[cid] = n
	}
	return m, nil
}

func (r *TutorialRepo) CountByMonth(n int) ([]int64, error) {
	sql := `SELECT COUNT(*) FROM tutorials WHERE status = ? AND created_at >= date('now','-'||?||' month') GROUP BY strftime('%Y-%m', created_at) ORDER BY created_at DESC`
	rows, err := r.db.Raw(sql, domain.TutorialStatusPublished, n).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]int64, 0, n)
	for rows.Next() {
		var x int64
		rows.Scan(&x)
		out = append(out, x)
	}
	return out, nil
}

func (r *TutorialRepo) SaveVersion(v *domain.TutorialVersion) error {
	return r.db.Create(v).Error
}
