package service

import (
	"errors"
	"time"
	"upcycle-hub/config"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
	apperr "upcycle-hub/pkg/errors"
	"upcycle-hub/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepo
	cfg      *config.JWTConfig
}

func NewAuthService(userRepo *repository.UserRepo, cfg *config.JWTConfig) *AuthService {
	return &AuthService{userRepo: userRepo, cfg: cfg}
}

type JWTClaims struct {
	UserID   uint64 `json:"uid"`
	Username string `json:"uname"`
	jwt.RegisteredClaims
}

func (s *AuthService) Register(username, email, password string) (*domain.User, error) {
	if !utils.IsValidEmail(email) {
		return nil, apperr.New(apperr.CodeValidation, "邮箱格式无效")
	}
	if !utils.IsValidPassword(password) {
		return nil, apperr.New(apperr.CodeValidation, "密码需6-32位且包含字母数字")
	}
	if len(username) < 3 || len(username) > 30 {
		return nil, apperr.New(apperr.CodeValidation, "用户名长度3-30字符")
	}
	if _, err := s.userRepo.GetByEmail(email); err == nil {
		return nil, apperr.ErrUserExists
	}
	if _, err := s.userRepo.GetByUsername(username); err == nil {
		return nil, apperr.New(apperr.CodeConflict, "用户名已存在")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "密码加密失败", err)
	}
	u := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Nickname:     username,
		Level:        domain.UserLevelNovice,
		Status:       1,
	}
	u.ComputeLevel()
	if err := s.userRepo.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *AuthService) Login(account, password string) (*domain.User, string, error) {
	var u *domain.User
	var err error
	if utils.IsValidEmail(account) {
		u, err = s.userRepo.GetByEmail(account)
	} else {
		u, err = s.userRepo.GetByUsername(account)
	}
	if err != nil {
		return nil, "", err
	}
	if u.Status != 1 {
		return nil, "", apperr.New(apperr.CodeForbidden, "账户已被禁用")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, "", apperr.ErrWrongPassword
	}
	token, err := s.GenerateToken(u)
	if err != nil {
		return nil, "", err
	}
	return u, token, nil
}

func (s *AuthService) GenerateToken(u *domain.User) (string, error) {
	claims := JWTClaims{
		UserID:   u.ID,
		Username: u.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.cfg.ExpireHour) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "upcycle-hub",
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(s.cfg.Secret))
}

func (s *AuthService) ParseToken(tokenStr string) (*JWTClaims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperr.ErrToken
		}
		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperr.New(apperr.CodeUnauthorized, "令牌已过期")
		}
		return nil, apperr.Wrap(apperr.CodeUnauthorized, "令牌无效", err)
	}
	claims, ok := t.Claims.(*JWTClaims)
	if !ok || !t.Valid {
		return nil, apperr.ErrUnauthorized
	}
	return claims, nil
}

func (s *AuthService) ResetPassword(id uint64, oldPwd, newPwd string) error {
	u, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPwd)); err != nil {
		return apperr.ErrWrongPassword
	}
	if !utils.IsValidPassword(newPwd) {
		return apperr.New(apperr.CodeValidation, "新密码需6-32位且包含字母数字")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternal, "密码加密失败", err)
	}
	return s.userRepo.UpdatePassword(id, string(hash))
}

func (s *AuthService) GetUserByID(id uint64) (*domain.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *AuthService) UpdateProfile(u *domain.User) error {
	old, err := s.userRepo.GetByID(u.ID)
	if err != nil {
		return err
	}
	if u.Nickname != "" {
		old.Nickname = u.Nickname
	}
	if u.Specialty != "" {
		old.Specialty = u.Specialty
	}
	if u.Bio != "" {
		old.Bio = u.Bio
	}
	if u.Avatar != "" {
		old.Avatar = u.Avatar
	}
	return s.userRepo.Update(old)
}
