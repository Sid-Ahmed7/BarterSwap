package main

import (
	"context"
	"database/sql"
)

type UserStore interface {
	CreateUser(ctx context.Context, r UserRequest) (User, error)
	GetUserByID(ctx context.Context, id int) (User, error)
	UpdateUser(ctx context.Context, id int, r UserRequest) (User, error)
	GetSkillsByUserID(ctx context.Context, userID int) ([]Skill, error)
	ReplaceSkills(ctx context.Context, userID int, skills []Skill) error
}

type ServiceStore interface {
	CreateService(ctx context.Context, providerID int, r ServiceRequest) (Service, error)
	GetServiceByID(ctx context.Context, id int) (Service, error)
	UpdateService(ctx context.Context, id int, r ServiceRequest) (Service, error)
	DeleteService(ctx context.Context, id int) error
	ListServices(ctx context.Context, filter ServiceListRequest) ([]Service, error)
	HasSkillsForCategory(ctx context.Context, userID int, categorie string) (bool, error)
}

type ExchangeStore interface {
	CreateExchange(ctx context.Context, req ExchangeRequest) (Exchange, error)
	GetExchangeByID(ctx context.Context, id int) (Exchange, error)
	ListExchanges(ctx context.Context, userID int, status string) ([]Exchange, error)
	HasActiveExchange(ctx context.Context, serviceID int) (bool, error)
	AcceptExchange(ctx context.Context, id int) (Exchange, error)
	RejectExchange(ctx context.Context, id int) (Exchange, error)
	CompleteExchange(ctx context.Context, id int) (Exchange, error)
	CancelExchange(ctx context.Context, id int) (Exchange, error)
}

type ReviewStore interface {
	CreateReview(ctx context.Context, exchangeID int, authorID int, req ReviewRequest) (Review, error)
	GetReviewsByUserID(ctx context.Context, userID int) ([]Review, error)
	GetReviewsByServiceID(ctx context.Context, serviceID int) ([]Review, error)
}

type StatsStore interface {
	GetUserStats(ctx context.Context, userID int) (UserStats, error)
}
type DB struct{ *sql.DB }
