// Copyright 2022 Teamgram Authors
// All rights reserved.
//
// Author: TeamgramIO (teamgram.io@gmail.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.

package config

import (
	"github.com/teamgram/proto/mtproto"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	mtproto.MTProtoConfig
	ChatClient       zrpc.RpcClientConf
	UserClient       zrpc.RpcClientConf
	MsgClient        zrpc.RpcClientConf
	DialogClient     zrpc.RpcClientConf
	SyncClient       zrpc.RpcClientConf
	MediaClient      zrpc.RpcClientConf
	AuthsessionClient zrpc.RpcClientConf
	IdgenClient      zrpc.RpcClientConf
	MessageClient    zrpc.RpcClientConf
	ChannelClient    zrpc.RpcClientConf
	BotClient        zrpc.RpcClientConf
	ReactionClient   zrpc.RpcClientConf
}
