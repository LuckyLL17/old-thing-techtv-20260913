package repository

import (
	"time"
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
	list := make([]*domain.Message, 0)
	q := r.db.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		userID, otherID, otherID, userID)
	if size > 0 {
		q = q.Offset((page - 1) * size).Limit(size)
	}
	err := q.Order("id ASC").Find(&list).Error
	return list, err
}

// ListConversations 返回与 userID 有过私信往来的会话，按最后一条消息时间倒序
func (r *MessageRepo) ListConversations(userID uint64) ([]*domain.Conversation, error) {
	list := make([]*domain.Conversation, 0)
	err := r.db.Raw(`
SELECT m.other_id, m.sender_id AS last_sender_id, m.content AS last_content, m.created_at AS last_at, COALESCE(unread, 0) AS unread
FROM (
  SELECT sender_id, CASE WHEN sender_id = ? THEN receiver_id ELSE sender_id END AS other_id, id, content, created_at
  FROM messages WHERE sender_id = ? OR receiver_id = ?
) m
JOIN (SELECT MAX(id) AS max_id FROM messages WHERE sender_id = ? OR receiver_id = ? GROUP BY CASE WHEN sender_id = ? THEN receiver_id ELSE sender_id END) latest
  ON m.id = latest.max_id
LEFT JOIN (
  SELECT sender_id, COUNT(*) AS unread FROM messages
  WHERE receiver_id = ? AND is_read = 0 GROUP BY sender_id
) u ON u.sender_id = m.other_id
ORDER BY m.id DESC`,
		userID, userID, userID, userID, userID, userID, userID).
		Scan(&list).Error
	return list, err
}

func (r *MessageRepo) MarkRead(userID, otherID uint64) error {
	now := time.Now()
	return r.db.Model(&domain.Message{}).
		Where("sender_id = ? AND receiver_id = ? AND is_read = ?", otherID, userID, false).
		Updates(map[string]interface{}{"is_read": true, "read_at": now}).Error
}

func (r *MessageRepo) UnreadCount(userID uint64) (int64, error) {
	var n int64
	err := r.db.Model(&domain.Message{}).
		Where("receiver_id = ? AND is_read = ?", userID, false).Count(&n).Error
	return n, err
}
