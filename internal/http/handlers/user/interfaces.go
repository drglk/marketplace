package userhandler

import "context"

const pkg = "userHandler/"

type AuthService interface {
	Register(ctx context.Context, login string, password string) error
}
