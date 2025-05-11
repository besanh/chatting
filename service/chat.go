package service

import (
	"context"
	"fmt"

	"github.com/besanh/chatting/common/constant"
	"github.com/besanh/chatting/model"
	"github.com/besanh/chatting/repository"
	log "github.com/besanh/logger/logging/slog"
	"github.com/uptrace/bun"
)

type (
	IChat interface {
		TxInsertChat(ctx context.Context, chat model.PinelineCreateChatRequest) (id string, err error)
		// GetChatById(ctx context.Context, chatId string) (model.Chat, error)
		// GetChats(ctx context.Context, filter model.Param) (int, []model.Chat, error)
		// UpdateChatById(ctx context.Context, chat model.Chat) error
		// DeleteChatById(ctx context.Context, chatId string) error
	}

	Chat struct {
		ChatRepo       repository.IChat
		ChatMemberRepo repository.IChatMember
	}
)

func NewChat(chatRepo repository.IChat, chatMemberRepo repository.IChatMember) IChat {
	return &Chat{
		ChatRepo:       chatRepo,
		ChatMemberRepo: chatMemberRepo,
	}
}

func (s *Chat) TxInsertChat(ctx context.Context, request model.PinelineCreateChatRequest) (id string, err error) {
	// Case 1: 1:1 chat, check if 2 members exist a conversation
	chatExist, err := s.ChatRepo.GetChatExist(ctx, repository.DBConn, model.ChatFilter{
		MemberIds: request.MemberIds,
	})
	if err != nil {
		log.Error(err)
		return
	} else if chatExist != nil {
		err = fmt.Errorf("chat already exists between %s and %s", request.MemberIds[0], request.MemberIds[1])
		log.Error(err)
		return
	}

	chat := &model.Chat{
		GBase:   model.InitPgBase(),
		Title:   request.Title,
		IsGroup: request.IsGroup,
	}

	chatMembers := make([]model.ChatMember, len(request.MemberIds))
	for i, id := range request.MemberIds {
		// TODO: at least 1 admin
		chatMembers[i] = model.ChatMember{
			GBase:  model.InitPgBase(),
			ChatId: chat.GetId(),
			UserId: id,
			Role:   constant.ROLE_MEMBER,
		}
	}

	if err = repository.DBConn.GetDB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) (err error) {
		if err = s.ChatRepo.TxInsert(ctx, tx, *chat); err != nil {
			log.Error(err)
			return
		}

		if err = s.ChatMemberRepo.TxBulkInsert(ctx, tx, chatMembers); err != nil {
			log.Error(err)
			return err
		}

		return
	}); err != nil {
		return
	}

	return chat.GetId(), err
}
