package dto

type RegisterReq struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginReq struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ResetPwdReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type UpdateProfileReq struct {
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Specialty string `json:"specialty"`
	Bio       string `json:"bio"`
}

type AuthResp struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

type PageReq struct {
	Page int `form:"page,default=1"`
	Size int `form:"size,default=10"`
}

type IdReq struct {
	ID uint64 `uri:"id" binding:"required,min=1"`
}

type Resp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}

type PageResp struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}
