package service

import "github.com/crazyfrankie/voidx/internal/repository/dao"

type ChatService struct {
	dao *dao.ChatDao
}

func NewChatService(dao *dao.ChatDao) *ChatService {
	return &ChatService{dao: dao}
}
