package repository

import (
	"upcycle-hub/internal/domain"
	apperr "upcycle-hub/pkg/errors"

	"gorm.io/gorm"
)

type StepRepo struct {
	db *gorm.DB
}

func NewStepRepo(db *gorm.DB) *StepRepo {
	return &StepRepo{db: db}
}

func (r *StepRepo) Create(s *domain.Step) error {
	err := r.db.Create(s).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "创建步骤失败", err)
	}
	return nil
}

func (r *StepRepo) BatchCreate(list []*domain.Step) error {
	if len(list) == 0 {
		return nil
	}
	err := r.db.Create(&list).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "批量创建步骤失败", err)
	}
	return nil
}

func (r *StepRepo) GetByID(id uint64) (*domain.Step, error) {
	s := &domain.Step{}
	err := r.db.First(s, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询步骤失败", err)
	}
	return s, nil
}

func (r *StepRepo) Update(s *domain.Step) error {
	err := r.db.Save(s).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "更新步骤失败", err)
	}
	return nil
}

func (r *StepRepo) Delete(id uint64) error {
	err := r.db.Delete(&domain.Step{}, id).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "删除步骤失败", err)
	}
	return nil
}

func (r *StepRepo) DeleteByTutorial(tid uint64) error {
	err := r.db.Where("tutorial_id = ?", tid).Delete(&domain.Step{}).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "删除步骤失败", err)
	}
	return nil
}

func (r *StepRepo) ListByTutorial(tid uint64) ([]*domain.Step, error) {
	var list []*domain.Step
	err := r.db.Where("tutorial_id = ?", tid).Order("step_order ASC").Find(&list).Error
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDB, "查询步骤列表失败", err)
	}
	return list, nil
}

func (r *StepRepo) UpdateOrder(tid uint64, order []uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, sid := range order {
			if err := tx.Model(&domain.Step{}).Where("id = ? AND tutorial_id = ?", sid, tid).
				Update("step_order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *StepRepo) ReorderAfterDelete(tid uint64) error {
	rows, err := r.db.Where("tutorial_id = ?", tid).Order("step_order ASC, id ASC").
		Select("id").Rows()
	if err != nil {
		return err
	}
	defer rows.Close()
	var ids []uint64
	for rows.Next() {
		var id uint64
		rows.Scan(&id)
		ids = append(ids, id)
	}
	return r.UpdateOrder(tid, ids)
}

func (r *StepRepo) ReplaceByTutorial(tid uint64, list []*domain.Step) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return apperr.Wrap(apperr.CodeDB, "启动事务失败", tx.Error)
	}
	if err := tx.Where("tutorial_id = ?", tid).Delete(&domain.Step{}).Error; err != nil {
		tx.Rollback()
		return apperr.Wrap(apperr.CodeDB, "重置步骤失败", err)
	}
	if len(list) > 0 {
		if err := tx.Create(&list).Error; err != nil {
			tx.Rollback()
			return apperr.Wrap(apperr.CodeDB, "批量创建步骤失败", err)
		}
	}
	if err := tx.Commit().Error; err != nil {
		return apperr.Wrap(apperr.CodeDB, "提交事务失败", err)
	}
	return nil
}
