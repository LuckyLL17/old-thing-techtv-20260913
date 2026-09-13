package repository

import (
	"upcycle-hub/internal/domain"

	"gorm.io/gorm"
)

type FollowRepo struct {
	db *gorm.DB
}

func NewFollowRepo(db *gorm.DB) *FollowRepo {
	return &FollowRepo{db: db}
}

func (r *FollowRepo) Follow(followerID, followingID uint64) error {
	f := &domain.Follow{FollowerID: followerID, FollowingID: followingID}
	return r.db.FirstOrCreate(f, domain.Follow{FollowerID: followerID, FollowingID: followingID}).Error
}

func (r *FollowRepo) Unfollow(followerID, followingID uint64) error {
	return r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(&domain.Follow{}).Error
}

func (r *FollowRepo) IsFollowing(followerID, followingID uint64) (bool, error) {
	var n int64
	err := r.db.Model(&domain.Follow{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&n).Error
	return n > 0, err
}

func (r *FollowRepo) Followers(userID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.Model(&domain.Follow{}).Where("following_id = ?", userID).Pluck("follower_id", &ids).Error
	return ids, err
}

func (r *FollowRepo) Following(userID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.Model(&domain.Follow{}).Where("follower_id = ?", userID).Pluck("following_id", &ids).Error
	return ids, err
}

func (r *FollowRepo) Counts(userID uint64) (followers, following int64, err error) {
	err = r.db.Model(&domain.Follow{}).Where("following_id = ?", userID).Count(&followers).Error
	if err != nil {
		return
	}
	err = r.db.Model(&domain.Follow{}).Where("follower_id = ?", userID).Count(&following).Error
	return
}

type MessageRepo struct {
	db *gorm.DB
}

func NewMessageRepo(db *gorm.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

func (r *MessageRepo) Send(m *domain.Message) error {
	return r.db.Create(m).Error
}

func (r *MessageRepo) List(userID, otherID uint64, page, size int) ([]*domain.Message, error) {
	var list []*domain.Message
	q := r.db.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		userID, otherID, otherID, userID)
	if size > 0 {
		q = q.Offset((page - 1) * size).Limit(size)
	}
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

func (r *MessageRepo) MarkRead(userID, otherID uint64) error {
	return r.db.Model(&domain.Message{}).
		Where("sender_id = ? AND receiver_id = ? AND is_read = ?", otherID, userID, false).
		Updates(map[string]interface{}{"is_read": true}).Error
}

func (r *MessageRepo) UnreadCount(userID uint64) (int64, error) {
	var n int64
	err := r.db.Model(&domain.Message{}).
		Where("receiver_id = ? AND is_read = ?", userID, false).Count(&n).Error
	return n, err
}
