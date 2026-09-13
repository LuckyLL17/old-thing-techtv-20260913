package service

import (
	"strings"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
)

type InteractionService struct {
	commentRepo  *repository.CommentRepo
	favoriteRepo *repository.FavoriteRepo
	attemptRepo  *repository.AttemptRepo
	followRepo   *repository.FollowRepo
	messageRepo  *repository.MessageRepo
	userRepo     *repository.UserRepo
	tutorialRepo *repository.TutorialRepo
	projectRepo  *repository.ProjectRepo
}

func NewInteractionService(cr *repository.CommentRepo, fr *repository.FavoriteRepo, ar *repository.AttemptRepo,
	flr *repository.FollowRepo, mr *repository.MessageRepo, ur *repository.UserRepo,
	tur *repository.TutorialRepo, pr *repository.ProjectRepo) *InteractionService {
	return &InteractionService{commentRepo: cr, favoriteRepo: fr, attemptRepo: ar, followRepo: flr,
		messageRepo: mr, userRepo: ur, tutorialRepo: tur, projectRepo: pr}
}

func (s *InteractionService) Comment(userID uint64, targetType string, targetID uint64, content string, parentID uint64) (*domain.Comment, error) {
	if content == "" {
		return nil, ErrValidation("评论内容不能为空")
	}
	if targetType != domain.CommentTypeTutorial && targetType != domain.CommentTypeProject {
		return nil, ErrValidation("评论目标类型无效")
	}
	c := &domain.Comment{
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
		ParentID:   parentID,
		Content:    content,
		Status:     1,
	}
	if err := s.commentRepo.Create(c); err != nil {
		return nil, err
	}
	if targetType == domain.CommentTypeTutorial {
		s.tutorialRepo.IncCounts(targetID, 0, 0, 1, 0)
	} else {
		s.projectRepo.IncComment(targetID, 1)
	}
	return s.commentRepo.GetByID(c.ID)
}

func (s *InteractionService) DeleteComment(id, userID uint64) error {
	c, err := s.commentRepo.GetByID(id)
	if err != nil {
		return err
	}
	if c.UserID != userID {
		return ErrForbidden("无权删除此评论")
	}
	if err := s.commentRepo.Delete(id); err != nil {
		return err
	}
	if c.TargetType == domain.CommentTypeTutorial {
		s.tutorialRepo.IncCounts(c.TargetID, 0, 0, -1, 0)
	} else {
		s.projectRepo.IncComment(c.TargetID, -1)
	}
	return nil
}

func (s *InteractionService) ListComments(targetType string, targetID uint64, page, size int) ([]*domain.Comment, int64, error) {
	return s.commentRepo.List(targetType, targetID, page, size)
}

func (s *InteractionService) ToggleFavorite(userID uint64, targetType string, targetID uint64) (bool, error) {
	ok, err := s.favoriteRepo.Exists(userID, targetType, targetID)
	if err != nil {
		return false, err
	}
	if ok {
		err = s.favoriteRepo.Delete(userID, targetType, targetID)
		if err == nil && targetType == domain.FavTypeTutorial {
			s.tutorialRepo.IncCounts(targetID, -1, 0, 0, 0)
		}
		return false, err
	}
	f := &domain.Favorite{UserID: userID, TargetType: targetType, TargetID: targetID}
	err = s.favoriteRepo.Create(f)
	if err == nil && targetType == domain.FavTypeTutorial {
		s.tutorialRepo.IncCounts(targetID, 1, 0, 0, 0)
	}
	return true, err
}

func (s *InteractionService) IsFavorite(userID uint64, targetType string, targetID uint64) (bool, error) {
	return s.favoriteRepo.Exists(userID, targetType, targetID)
}

func (s *InteractionService) ListFavorites(userID uint64, targetType string, page, size int) ([]*domain.Favorite, int64, error) {
	return s.favoriteRepo.ListByUser(userID, targetType, page, size)
}

func (s *InteractionService) Follow(followerID, followingID uint64) error {
	if followerID == followingID {
		return ErrValidation("不能关注自己")
	}
	return s.followRepo.Follow(followerID, followingID)
}

func (s *InteractionService) Unfollow(followerID, followingID uint64) error {
	return s.followRepo.Unfollow(followerID, followingID)
}

func (s *InteractionService) IsFollowing(followerID, followingID uint64) (bool, error) {
	return s.followRepo.IsFollowing(followerID, followingID)
}

func (s *InteractionService) FollowCounts(userID uint64) (int64, int64, error) {
	return s.followRepo.Counts(userID)
}

const maxMessageLen = 2000

func (s *InteractionService) SendMessage(senderID, receiverID uint64, content string) (*domain.Message, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ErrValidation("消息内容不能为空")
	}
	if len([]rune(content)) > maxMessageLen {
		return nil, ErrValidation("消息内容过长")
	}
	if senderID == receiverID {
		return nil, ErrValidation("不能给自己发消息")
	}
	receiver, err := s.userRepo.GetByID(receiverID)
	if err != nil {
		return nil, err
	}
	if receiver.Status != 1 {
		return nil, ErrValidation("该用户当前无法接收消息")
	}
	m := &domain.Message{SenderID: senderID, ReceiverID: receiverID, Content: content}
	if err := s.messageRepo.Send(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ListMessages 返回与 otherID 的完整聊天历史（时间正序），并把对方发来的未读标记为已读
func (s *InteractionService) ListMessages(userID, otherID uint64) ([]*domain.Message, error) {
	if err := s.messageRepo.MarkRead(userID, otherID); err != nil {
		return nil, err
	}
	return s.messageRepo.List(userID, otherID, 1, 0)
}

func (s *InteractionService) ListConversations(userID uint64) ([]*domain.Conversation, error) {
	list, err := s.messageRepo.ListConversations(userID)
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(list))
	for _, c := range list {
		ids = append(ids, c.OtherID)
	}
	users, err := s.userRepo.GetByIDs(ids)
	if err != nil {
		return nil, err
	}
	for _, c := range list {
		if u, ok := users[c.OtherID]; ok {
			c.User = &domain.ConversationUser{
				ID:       u.ID,
				Username: u.Username,
				Nickname: u.Nickname,
				Avatar:   u.Avatar,
			}
		}
	}
	return list, nil
}

func (s *InteractionService) UnreadCount(userID uint64) (int64, error) {
	return s.messageRepo.UnreadCount(userID)
}

func (s *InteractionService) GetUserProfile(userID uint64) (*domain.User, error) {
	return s.userRepo.GetByID(userID)
}

func (s *InteractionService) MarkAttempt(userID, tutorialID uint64, completed bool, note string) error {
	a, err := s.attemptRepo.GetByUserAndTutorial(userID, tutorialID)
	if err != nil {
		return err
	}
	if a == nil {
		a = &domain.Attempt{UserID: userID, TutorialID: tutorialID, Completed: completed, Note: note}
		err = s.attemptRepo.Create(a)
		if err == nil {
			s.tutorialRepo.IncCounts(tutorialID, 0, 1, 0, 0)
		}
		return err
	}
	a.Completed = completed
	if note != "" {
		a.Note = note
	}
	return s.attemptRepo.Update(a)
}

func (s *InteractionService) ListAttempts(userID uint64, page, size int) ([]*domain.Attempt, int64, error) {
	return s.attemptRepo.ListByUser(userID, page, size)
}
