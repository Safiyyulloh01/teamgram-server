// Copyright 2022 Teamgram Authors
// All rights reserved.
// Author: TeamgramIO (teamgram.io@gmail.com)

package core

import (
	"context"

	"github.com/teamgram/teamgram-server/app/bff/channels/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/marmota/pkg/metadata"
)

type ChannelsCore struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	MD *metadata.RpcMetadata
}

func New(ctx context.Context, svcCtx *svc.ServiceContext) *ChannelsCore {
	return &ChannelsCore{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		MD:     metadata.RpcMetadataFromIncoming(ctx),
	}
}
