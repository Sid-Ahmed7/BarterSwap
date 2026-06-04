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
}

type DB struct{ *sql.DB }
