package router

import "plane_war/internal/api"

func (r RouterGroup) AuthRouter() {
	r.POST("auth/login", api.Login)
	r.POST("auth/refresh_token", api.RefreshToken)
	r.POST("auth/register", api.Register)
}
