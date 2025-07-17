package req

type LoginReq struct {
	Email    string `json:"email" binding:"required,max=100,min=5,email"`
	Password string `json:"password" binding:"required,max=16,min=8"`
}

type UpdatePasswdReq struct {
	Password string `json:"password" binding:"required,max=16,min=8"`
}

type UpdateNameReq struct {
	Name string `json:"name"`
}

type UpdateAvatarReq struct {
	Avatar string `json:"avatar"`
}
