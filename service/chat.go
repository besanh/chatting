package service

import (
	"context"

	"github.com/besanh/chatting/common/constant"
	"github.com/besanh/chatting/model"
	"github.com/besanh/chatting/repository"
	log "github.com/besanh/logger/logging/slog"
	"github.com/uptrace/bun"
)

type (
	IChat interface {
		InsertChat(ctx context.Context, chat model.PinelineCreateChatRequest) (id string, err error)
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

func (s *Chat) InsertChat(ctx context.Context, request model.PinelineCreateChatRequest) (id string, err error) {
	chat := &model.Chat{
		GBase:   model.InitPgBase(),
		Title:   request.Title,
		IsGroup: request.IsGroup.Bool,
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

	err = s.ChatRepo.Insert(ctx, repository.DBConn, *chat)
	if err != nil {
		log.Error(err)
		return
	}

	return chat.GetId(), err
}
