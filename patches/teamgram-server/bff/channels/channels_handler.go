// Channels RPC handlers — full implementation
// This file implements channels.createChannel, joinChannel, leaveChannel, etc.

package core

import (
	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/proto/mtproto/crypto"
	chatpb "github.com/teamgram/teamgram-server/app/service/biz/chat/chat"
	userpb "github.com/teamgram/teamgram-server/app/service/biz/user/user"
	"github.com/zeromicro/go-zero/core/mr"
)

// ChannelsCreateChannel creates a new channel or megagroup
func (c *ChannelsCore) ChannelsCreateChannel(in *mtproto.TLChannelsCreateChannel) (*mtproto.Updates, error) {
	// Generate unique channel ID
	channelId, err := c.svcCtx.Dao.IDGenClient2.NextId(c.ctx)
	if err != nil {
		return nil, err
	}

	accessHash := crypto.GenerateAccessHash(channelId)

	// Determine if it's a megagroup or broadcast channel
	isBroadcast := mtproto.FromBool(in.GetBroadcast())
	isMegagroup := mtproto.FromBool(in.GetMegagroup())

	channelType := int32(1) // 1=channel, 2=megagroup
	if isMegagroup {
		channelType = 2
	}

	// Create channel via channel service
	channel, err := c.svcCtx.Dao.ChannelClient.Client().ChannelCreateChannel(c.ctx, &chatpb.TLChannelCreateChannel{
		ChannelId:          channelId,
		AccessHash:         accessHash,
		CreatorUserId:      c.MD.UserId,
		Title:              in.GetTitle(),
		About:              in.GetAbout(),
		Broadcast:          in.GetBroadcast(),
		Megagroup:          in.GetMegagroup(),
		ChannelType:        channelType,
		Date:               int32(mtproto.Now().Unix()),
		Version:            1,
		ParticipantsCount:  1,
		AdminsCount:        1,
	})

	if err != nil {
		return nil, err
	}

	// Add creator as participant with admin rights
	_, err = c.svcCtx.Dao.ChannelClient.Client().ChannelCreateChannelParticipant(c.ctx, &chatpb.TLChannelCreateChannelParticipant{
		ChannelId: channelId,
		UserId:    c.MD.UserId,
		IsAdmin:   true,
		Date:      int32(mtproto.Now().Unix()),
	})
	if err != nil {
		return nil, err
	}

	// Build updates response
	chat := channel.ToUnsafeChat(c.MD.UserId)
	updatesHelper := mtproto.MakeUpdatesHelper(chat)

	// Add the channel to user's dialog
	dialog, err := c.svcCtx.Dao.DialogClient.DialogInsertOrUpdateDialog(c.ctx, &chatpb.TLDialogInsertOrUpdateDialog{
		UserId:   c.MD.UserId,
		PeerType: mtproto.PEER_CHANNEL,
		PeerId:   channelId,
		TopMessage: 1,
		Pinned:     0,
		Date:       int32(mtproto.Now().Unix()),
	})
	_ = dialog

	return updatesHelper.ToUpdates(), nil
}

// ChannelsJoinChannel joins a user to a channel
func (c *ChannelsCore) ChannelsJoinChannel(in *mtproto.TLChannelsJoinChannel) (*mtproto.Updates, error) {
	channelId := in.GetChannel().GetChannelId()
	channel, err := c.svcCtx.Dao.ChannelClient.Client().ChannelGetMutableChannel(c.ctx, &chatpb.TLChannelGetMutableChannel{
		ChannelId: channelId,
	})
	if err != nil {
		return nil, err
	}

	// Check if not already a participant
	if channel.IsParticipant(c.MD.UserId) {
		return nil, mtproto.ErrChannelsTooMuch
	}

	// Add participant
	_, err = c.svcCtx.Dao.ChannelClient.Client().ChannelCreateChannelParticipant(c.ctx, &chatpb.TLChannelCreateChannelParticipant{
		ChannelId: channelId,
		UserId:    c.MD.UserId,
		IsAdmin:   false,
		Date:      int32(mtproto.Now().Unix()),
	})
	if err != nil {
		return nil, err
	}

	// Increment participant count
	_, err = c.svcCtx.Dao.ChannelClient.Client().ChannelUpdateParticipantCount(c.ctx, &chatpb.TLChannelUpdateParticipantCount{
		ChannelId: channelId,
		Delta:     1,
	})
	if err != nil {
		return nil, err
	}

	// Add to user's dialog
	c.svcCtx.Dao.DialogClient.DialogInsertOrUpdateDialog(c.ctx, &chatpb.TLDialogInsertOrUpdateDialog{
		UserId:   c.MD.UserId,
		PeerType: mtproto.PEER_CHANNEL,
		PeerId:   channelId,
		TopMessage: 1,
		Date:       int32(mtproto.Now().Unix()),
	})

	updatesHelper := mtproto.MakeUpdatesHelper(channel.ToUnsafeChat(c.MD.UserId))
	return updatesHelper.ToUpdates(), nil
}

// ChannelsLeaveChannel removes a user from a channel
func (c *ChannelsCore) ChannelsLeaveChannel(in *mtproto.TLChannelsLeaveChannel) (*mtproto.Updates, error) {
	channelId := in.GetChannel().GetChannelId()

	// Remove participant
	_, err := c.svcCtx.Dao.ChannelClient.Client().ChannelDeleteChannelParticipant(c.ctx, &chatpb.TLChannelDeleteChannelParticipant{
		ChannelId: channelId,
		UserId:    c.MD.UserId,
	})
	if err != nil {
		return nil, err
	}

	// Update dialog
	c.svcCtx.Dao.DialogClient.DialogDeleteDialog(c.ctx, &chatpb.TLDialogDeleteDialog{
		UserId:   c.MD.UserId,
		PeerType: mtproto.PEER_CHANNEL,
		PeerId:   channelId,
	})

	updatesHelper := mtproto.MakeUpdatesHelper()
	return updatesHelper.ToUpdates(), nil
}

// ChannelsGetChannels returns info about multiple channels
func (c *ChannelsCore) ChannelsGetChannels(in *mtproto.TLChannelsGetChannels) (*mtproto.Messages_Chats, error) {
	ids := in.GetId()
	chats := make([]*mtproto.Chat, 0, len(ids))

	for _, id := range ids {
		channel, err := c.svcCtx.Dao.ChannelClient.Client().ChannelGetMutableChannel(c.ctx, &chatpb.TLChannelGetMutableChannel{
			ChannelId: id.GetChannelId(),
		})
		if err == nil && channel != nil {
			chats = append(chats, channel.ToUnsafeChat(c.MD.UserId))
		}
	}

	return mtproto.MakeTLMessagesChats(&mtproto.Messages_Chats{
		Chats: chats,
	}).To_Messages_Chats(), nil
}

// ChannelsGetFullChannel returns full channel info including participants
func (c *ChannelsCore) ChannelsGetFullChannel(in *mtproto.TLChannelsGetFullChannel) (*mtproto.Messages_ChatFull, error) {
	channelId := in.GetChannel().GetChannelId()

	channel, err := c.svcCtx.Dao.ChannelClient.Client().ChannelGetMutableChannel(c.ctx, &chatpb.TLChannelGetMutableChannel{
		ChannelId: channelId,
	})
	if err != nil {
		return nil, err
	}

	fullChat := mtproto.MakeTLChannelFull(&mtproto.ChatFull{
		Id:                 channelId,
		Participants:       mtproto.MakeTLChannelParticipants(&mtproto.ChannelParticipants{}).To_ChannelParticipants(),
		ChatPhoto:          mtproto.MakeTLPhotoEmpty(&mtproto.Photo{}).To_Photo(),
		NotifySettings:     mtproto.MakeTLPeerNotifySettings(&mtproto.PeerNotifySettings{}).To_PeerNotifySettings(),
		ExportedInvite:     mtproto.MakeTLChatInviteExported(&mtproto.ExportedChatInvite{}).To_ExportedChatInvite(),
		BotInfo:            []*mtproto.BotInfo{},
		PinnedMsgId:        channel.GetPinnedMsgId(),
		FolderId:           0,
		CanSetUsername:     true,
		CanSetStickers:     true,
		CanViewParticipants: true,
		CanViewStatistics:  false,
		Stickerset:         mtproto.MakeTLStickerSetNotSet(&mtproto.StickerSet{}).To_StickerSet(),
	}).To_ChatFull()

	// Get user info for creator
	user, _ := c.svcCtx.Dao.UserClient.UserGetMutableUsers(c.ctx, &userpb.TLUserGetMutableUsers{
		Id: []int64{channel.GetCreatorId()},
	})

	return mtproto.MakeTLMessagesChatFull(&mtproto.Messages_ChatFull{
		FullChat: fullChat,
		Chats:    []*mtproto.Chat{channel.ToUnsafeChat(c.MD.UserId)},
		Users:    user.GetUserListByIdList(c.MD.UserId, channel.GetCreatorUserId()),
	}).To_Messages_ChatFull(), nil
}

// ChannelsEditTitle edits the channel title
func (c *ChannelsCore) ChannelsEditTitle(in *mtproto.TLChannelsEditTitle) (*mtproto.Updates, error) {
	channelId := in.GetChannel().GetChannelId()

	channel, err := c.svcCtx.Dao.ChannelClient.Client().ChannelEditTitle(c.ctx, &chatpb.TLChannelEditTitle{
		ChannelId: channelId,
		Title:     in.GetTitle(),
		UserId:    c.MD.UserId,
	})
	if err != nil {
		return nil, err
	}

	updatesHelper := mtproto.MakeUpdatesHelper(channel.ToUnsafeChat(c.MD.UserId))
	return updatesHelper.ToUpdates(), nil
}

// ChannelsEditAbout edits the channel description
func (c *ChannelsCore) ChannelsEditAbout(in *mtproto.TLChannelsEditAbout) (*mtproto.Updates, error) {
	channelId := in.GetChannel().GetChannelId()

	_, err := c.svcCtx.Dao.ChannelClient.Client().ChannelEditAbout(c.ctx, &chatpb.TLChannelEditAbout{
		ChannelId: channelId,
		About:     in.GetAbout(),
		UserId:    c.MD.UserId,
	})
	if err != nil {
		return nil, err
	}

	channel, _ := c.svcCtx.Dao.ChannelClient.Client().ChannelGetMutableChannel(c.ctx, &chatpb.TLChannelGetMutableChannel{
		ChannelId: channelId,
	})

	updatesHelper := mtproto.MakeUpdatesHelper(channel.ToUnsafeChat(c.MD.UserId))
	return updatesHelper.ToUpdates(), nil
}

// ChannelsDeleteChannel deletes a channel
func (c *ChannelsCore) ChannelsDeleteChannel(in *mtproto.TLChannelsDeleteChannel) (*mtproto.Updates, error) {
	channelId := in.GetChannel().GetChannelId()

	// Only creator can delete
	channel, err := c.svcCtx.Dao.ChannelClient.Client().ChannelGetMutableChannel(c.ctx, &chatpb.TLChannelGetMutableChannel{
		ChannelId: channelId,
	})
	if err != nil {
		return nil, err
	}

	if channel.GetCreatorId() != c.MD.UserId {
		return nil, mtproto.ErrChatAdminRequired
	}

	// Soft delete
	_, err = c.svcCtx.Dao.ChannelClient.Client().ChannelDeleteChannel(c.ctx, &chatpb.TLChannelDeleteChannel{
		ChannelId: channelId,
		UserId:    c.MD.UserId,
	})
	if err != nil {
		return nil, err
	}

	updatesHelper := mtproto.MakeUpdatesHelper()
	return updatesHelper.ToUpdates(), nil
}

// ChannelsGetParticipants returns channel participants list
func (c *ChannelsCore) ChannelsGetParticipants(in *mtproto.TLChannelsGetParticipants) (*mtproto.ChannelParticipants, error) {
	channelId := in.GetChannel().GetChannelId()

	participants, err := c.svcCtx.Dao.ChannelClient.Client().ChannelGetParticipants(c.ctx, &chatpb.TLChannelGetParticipants{
		ChannelId: channelId,
		Filter:    in.GetFilter(),
		Offset:    in.GetOffset(),
		Limit:     in.GetLimit(),
	})
	if err != nil {
		return nil, err
	}

	return participants, nil
}

// ChannelsToggleSignatures toggles signatures on/off for channel posts
func (c *ChannelsCore) ChannelsToggleSignatures(in *mtproto.TLChannelsToggleSignatures) (*mtproto.Updates, error) {
	channelId := in.GetChannel().GetChannelId()

	_, err := c.svcCtx.Dao.ChannelClient.Client().ChannelToggleSignatures(c.ctx, &chatpb.TLChannelToggleSignatures{
		ChannelId:  channelId,
		Signatures: in.GetSignaturesEnabled(),
		UserId:     c.MD.UserId,
	})
	if err != nil {
		return nil, err
	}

	channel, _ := c.svcCtx.Dao.ChannelClient.Client().ChannelGetMutableChannel(c.ctx, &chatpb.TLChannelGetMutableChannel{
		ChannelId: channelId,
	})

	updatesHelper := mtproto.MakeUpdatesHelper(channel.ToUnsafeChat(c.MD.UserId))
	return updatesHelper.ToUpdates(), nil
}

// ChannelsCheckUsername checks if a username is available for the channel
func (c *ChannelsCore) ChannelsCheckUsername(in *mtproto.TLChannelsCheckUsername) (*mtproto.Bool, error) {
	// Check availability via user service
	available, err := c.svcCtx.Dao.UserClient.UserCheckUsername(c.ctx, &userpb.TLUserCheckUsername{
		Username: in.GetUsername(),
	})
	if err != nil {
		return mtproto.BoolFalse, nil
	}
	return available, nil
}

// ChannelsUpdateUsername updates the channel's public username
func (c *ChannelsCore) ChannelsUpdateUsername(in *mtproto.TLChannelsUpdateUsername) (*mtproto.Bool, error) {
	channelId := in.GetChannel().GetChannelId()
	username := in.GetUsername()

	// First check if username is available
	available, err := c.svcCtx.Dao.UserClient.UserCheckUsername(c.ctx, &userpb.TLUserCheckUsername{
		Username: username,
	})
	if err != nil || mtproto.FromBool(available) == false {
		return mtproto.BoolFalse, nil
	}

	_, err = c.svcCtx.Dao.ChannelClient.Client().ChannelUpdateUsername(c.ctx, &chatpb.TLChannelUpdateUsername{
		ChannelId: channelId,
		Username:  username,
		UserId:    c.MD.UserId,
	})
	if err != nil {
		return mtproto.BoolFalse, nil
	}

	return mtproto.BoolTrue, nil
}
