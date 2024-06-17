package main

import (
	"fmt"

	"google.golang.org/grpc/codes"
)

type Err struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e Err) Message(msg string) Err {
	ne := Err{
		Code: e.Code,
	}
	ne.Msg = msg
	return ne
}

func (e Err) Error() string {
	return fmt.Sprintf(`{"code": %d, "msg": "%s"}`, e.Code, e.Msg)
}

func NewErr(code int, msg string) *Err {
	return &Err{Code: code, Msg: msg}
}

var (
	ErrDb = NewErr(int(codes.Internal), "db err")
	ErrInvalidArgs = NewErr(int(codes.InvalidArgument), "invalid args")
)
