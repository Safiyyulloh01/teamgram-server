// Copyright 2022 Teamgram Authors
// All rights reserved.
//
// Author: TeamgramIO (teamgram.io@gmail.com)

package dao

import (
	"github.com/teamgram/marmota/pkg/net/rpcx"
	"github.com/teamgram/teamgram-server/app/bff/channels/internal/config"
	sync_client "github.com/teamgram/teamgram-server/app/interface/session/client"
	user_client "github.com/teamgram/teamgram-server/app/service/biz/user/client"
	chat_client "github.com/teamgram/teamgram-server/app/service/biz/chat/client"
	msg_client "github.com/teamgram/teamgram-server/app/messenger/msg/client"
	dialog_client "github.com/teamgram/teamgram-server/app/service/biz/dialog/client"
	media_client "github.com/teamgram/teamgram-server/app/service/media/client"
	authsession_client "github.com/teamgram/teamgram-server/app/service/authsession/client"
	idgen_client "github.com/teamgram/teamgram-server/app/service/idgen/client"
	message_client "github.com/teamgram/teamgram-server/app/service/biz/message/client"
	channel_client "github.com/teamgram/teamgram-server/app/service/channel/client"
	"github.com/zeromicro/go-zero/zrpc"
)

type Dao struct {
	user_client.UserClient
	ChatClient       *chat_client.ChatClientHelper
	msg_client.MsgClient
	sync_client.SyncClient
	media_client.MediaClient
	dialog_client.DialogClient
	authsession_client.AuthsessionClient
	idgen_client.IDGenClient2
	message_client.MessageClient
	ChannelClient    *channel_client.ChannelClientHelper
}

func New(c config.Config) *Dao {
	return &Dao{
		UserClient:       user_client.NewUserClient(rpcx.GetCachedRpcClient(c.UserClient)),
		ChatClient:       chat_client.NewChatClientHelper(rpcx.GetCachedRpcClient(c.ChatClient)),
		MsgClient:        msg_client.NewMsgClient(rpcx.GetCachedRpcClient(c.MsgClient)),
		SyncClient:       sync_client.NewSyncClient(rpcx.GetCachedRpcClient(c.SyncClient)),
		MediaClient:      media_client.NewMediaClient(rpcx.GetCachedRpcClient(c.MediaClient)),
		DialogClient:     dialog_client.NewDialogClient(rpcx.GetCachedRpcClient(c.DialogClient)),
		AuthsessionClient: authsession_client.NewAuthsessionClient(rpcx.GetCachedRpcClient(c.AuthsessionClient)),
		IDGenClient2:     idgen_client.NewIDGenClient2(rpcx.GetCachedRpcClient(c.IdgenClient)),
		MessageClient:    message_client.NewMessageClient(rpcx.GetCachedRpcClient(c.MessageClient)),
		ChannelClient:    channel_client.NewChannelClientHelper(rpcx.GetCachedRpcClient(c.ChannelClient)),
	}
}
