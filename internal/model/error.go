package model

import "errors"

var ErrUserNameOrPasswordFailed = errors.New("用户名或密码错误")
var ErrEmailAleadyExist = errors.New("邮箱已存在")
var ErrEmailCodeNotEqual = errors.New("邮箱验证码不正确")
