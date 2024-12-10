package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aeilang/urlshortener/internal/model"
	"github.com/aeilang/urlshortener/internal/repo"
)

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(passsword, hash string) bool
}

type UserCacher interface {
	GetEmailCode(ctx context.Context, email string) (string, error)
	SetEmailCode(ctx context.Context, email, emailCode string) error
}

type EmailSender interface {
	Send(email, emailCode string) error
}

type NumberRandomer interface {
	Generate() string
}

type JWTer interface {
	Generate(email string, userID int) (string, error)
}

type UserService struct {
	querier        repo.Querier
	passwordHasher PasswordHasher
	jwter          JWTer
	userCacher     UserCacher
	emailSender    EmailSender
	numberRandomer NumberRandomer
}

func NewUserService(db *sql.DB, p PasswordHasher, j JWTer, u UserCacher, e EmailSender, n NumberRandomer) *UserService {
	return &UserService{
		querier:        repo.New(db),
		passwordHasher: p,
		jwter:          j,
		userCacher:     u,
		emailSender:    e,
		numberRandomer: n,
	}
}

func (s *UserService) Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.querier.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %v", err)
	}

	if !s.passwordHasher.ComparePassword(req.Password, user.PasswordHash) {
		return nil, model.ErrUserNameOrPasswordFailed
	}

	accessToken, err := s.jwter.Generate(user.Email, int(user.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to genrate access token: %v", err)
	}

	return &model.LoginResponse{
		AccessToken: accessToken,
		Username:    user.Username,
		Email:       user.Email,
	}, nil
}

func (s *UserService) IsEmailAvaliable(ctx context.Context, email string) error {
	isAvaliable, err := s.querier.IsEmailAvaliable(ctx, email)
	if err != nil {
		return err
	}

	if !isAvaliable {
		return model.ErrEmailAleadyExist
	}

	return nil
}

func (s *UserService) Register(ctx context.Context, req model.RegisterReqeust) (*model.LoginResponse, error) {
	// 判断emailCode是否正确
	emailCode, err := s.userCacher.GetEmailCode(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get emailCode from cache: %v", err)
	}
	if emailCode != req.EmailCode {
		return nil, model.ErrEmailCodeNotEqual
	}

	// hash密码
	hash, err := s.passwordHasher.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// 插入数据库
	row, err := s.querier.CreateUser(ctx, repo.CreateUserParams{
		Username:     req.Username,
		PasswordHash: hash,
		Email:        req.Email,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// 生成access token
	accessToken, err := s.jwter.Generate(row.Email, int(row.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %v", err)
	}

	return &model.LoginResponse{
		AccessToken: accessToken,
		Username:    req.Username,
		Email:       req.Email,
	}, nil
}

func (s *UserService) SendEmailCode(ctx context.Context, email string) error {
	emailCode := s.numberRandomer.Generate()

	if err := s.emailSender.Send(email, emailCode); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	if err := s.userCacher.SetEmailCode(ctx, email, emailCode); err != nil {
		return fmt.Errorf("failed to set emailcode in cache: %v", err)
	}

	return nil
}

func (s *UserService) ResetPassword(ctx context.Context, req model.ForgetPasswordReqeust) (*model.LoginResponse, error) {
	emailCode, err := s.userCacher.GetEmailCode(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if emailCode != req.EmailCode {
		return nil, model.ErrEmailCodeNotEqual
	}

	hash, err := s.passwordHasher.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 更新数据库
	user, err := s.querier.UpdatePasswordByEmail(ctx, repo.UpdatePasswordByEmailParams{
		PasswordHash: hash,
		UpdatedAt:    time.Now(),
		Email:        req.Email,
	})
	if err != nil {
		return nil, err
	}

	accessToken, err := s.jwter.Generate(user.Email, int(user.ID))
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		AccessToken: accessToken,
		Email:       user.Email,
		Username:    user.Username,
	}, nil
}
