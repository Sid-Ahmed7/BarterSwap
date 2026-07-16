package store

import (
	"context"
	"database/sql"

	"barterswap/internal/model"
)

type UserStore interface {
	CreateUser(ctx context.Context, r model.UserRequest) (model.User, error)
	GetUserByID(ctx context.Context, id int) (model.User, error)
	UpdateUser(ctx context.Context, id int, r model.UserRequest) (model.User, error)
	GetSkillsByUserID(ctx context.Context, userID int) ([]model.Skill, error)
	ReplaceSkills(ctx context.Context, userID int, skills []model.Skill) error
}

type ServiceStore interface {
	CreateService(ctx context.Context, providerID int, r model.ServiceRequest) (model.Service, error)
	GetServiceByID(ctx context.Context, id int) (model.Service, error)
	UpdateService(ctx context.Context, id int, r model.ServiceRequest) (model.Service, error)
	DeleteService(ctx context.Context, id int) error
	ListServices(ctx context.Context, filter model.ServiceListRequest) ([]model.Service, error)
	HasSkillsForCategory(ctx context.Context, userID int, categorie string) (bool, error)
}

type ExchangeStore interface {
	CreateExchange(ctx context.Context, req model.ExchangeRequest) (model.Exchange, error)
	GetExchangeByID(ctx context.Context, id int) (model.Exchange, error)
	ListExchanges(ctx context.Context, userID int, status string) ([]model.Exchange, error)
	HasActiveExchange(ctx context.Context, serviceID int) (bool, error)
	AcceptExchange(ctx context.Context, id int) (model.Exchange, error)
	RejectExchange(ctx context.Context, id int) (model.Exchange, error)
	CompleteExchange(ctx context.Context, id int) (model.Exchange, error)
	CancelExchange(ctx context.Context, id int) (model.Exchange, error)
}

type ReviewStore interface {
	CreateReview(ctx context.Context, exchangeID int, authorID int, req model.ReviewRequest) (model.Review, error)
	GetReviewsByUserID(ctx context.Context, userID int) ([]model.Review, error)
	GetReviewsByServiceID(ctx context.Context, serviceID int) ([]model.Review, error)
}

type StatsStore interface {
	GetUserStats(ctx context.Context, userID int) (model.UserStats, error)
}

type DB struct{ *sql.DB }