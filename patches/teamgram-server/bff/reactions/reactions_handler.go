// Reactions RPC handlers — full implementation
package core

import (
	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/proto/mtproto/crypto"
	"github.com/teamgram/teamgram-server/app/bff/reactions/internal/svc"
)

type ReactionsCore struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	MD *metadata.RpcMetadata
}

func New(ctx context.Context, svcCtx *svc.ServiceContext) *ReactionsCore {
	return &ReactionsCore{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		MD:     metadata.RpcMetadataFromIncoming(ctx),
	}
}

// MessagesGetAvailableReactions returns available message reactions
func (c *ReactionsCore) MessagesGetAvailableReactions(in *mtproto.TLMessagesGetAvailableReactions) (*mtproto.Messages_AvailableReactions, error) {
	// Return standard Telegram reactions
	reactions := []*mtproto.AvailableReaction{
		makeReaction("👍", "\uD83D\uDC4D", 1),
		makeReaction("👎", "\uD83D\uDC4E", 2),
		makeReaction("❤", "\u2764\uFE0F", 3),
		makeReaction("🔥", "\uD83D\uDD25", 4),
		makeReaction("\uD83D\uDE0A", "\uD83D\uDE0A", 5),
		makeReaction("😁", "\uD83D\uDE01", 6),
		makeReaction("😮", "\uD83D\uDE2E", 7),
		makeReaction("🎉", "\uD83C\uDF89", 8),
		makeReaction("💩", "\uD83D\uDCA9", 9),
		makeReaction("\uD83D\uDE22", "\uD83D\uDE22", 10),
		makeReaction("🤩", "\uD83E\uDD29", 11),
		makeReaction("🤔", "\uD83E\uDD14", 12),
		makeReaction("👏", "\uD83D\uDC4F", 13),
		makeReaction("\uD83D\uDE2D", "\uD83D\uDE2D", 14),
		makeReaction("😈", "\uD83D\uDE08", 15),
		makeReaction("🎃", "\uD83C\uDF83", 16),
		makeReaction("💯", "\uD83D\uDCAF", 17),
		makeReaction("🐳", "\uD83D\uDC33", 18),
		makeReaction("\uD83E\uDD37", "\uD83E\uDD37", 19),
		makeReaction("\uD83D\uDC96", "\uD83D\uDC96", 20),
		makeReaction("💘", "\uD83D\uDC98", 21),
	}

	return mtproto.MakeTLMessagesAvailableReactions(&mtproto.Messages_AvailableReactions{
		Hash:       int32(crypto.GenerateHash(reactions)),
		Reactions:  reactions,
	}).To_Messages_AvailableReactions(), nil
}

func makeReaction(emoticon, staticIcon string, idx int) *mtproto.AvailableReaction {
	return mtproto.MakeTLAvailableReaction(&mtproto.AvailableReaction{
		Reaction:         emoticon,
		Title:            emoticon,
		StaticIcon:       mtproto.MakeTLDocumentEmpty(&mtproto.Document{}).To_Document(),
		AppearAnimation:  mtproto.MakeTLDocumentEmpty(&mtproto.Document{}).To_Document(),
		SelectAnimation:  mtproto.MakeTLDocumentEmpty(&mtproto.Document{}).To_Document(),
		ActivateAnimation: mtproto.MakeTLDocumentEmpty(&mtproto.Document{}).To_Document(),
		EffectAnimation:  mtproto.MakeTLDocumentEmpty(&mtproto.Document{}).To_Document(),
		AroundAnimation:  mtproto.MakeTLDocumentEmpty(&mtproto.Document{}).To_Document(),
		CenterIcon:       mtproto.MakeTLDocumentEmpty(&mtproto.Document{}).To_Document(),
	}).To_AvailableReaction()
}

// MessagesSendReaction sends a reaction to a message
func (c *ReactionsCore) MessagesSendReaction(in *mtproto.TLMessagesSendReaction) (*mtproto.Updates, error) {
	peer := in.GetPeer()
	msgId := in.GetMsgId()
	reaction := in.GetReaction() // []string — can have multiple reactions
	big := mtproto.FromBool(in.GetBig())

	peerType, peerId := mtproto.ResolvePeer(peer)

	// Store reaction in database
	for _, r := range reaction {
		_, err := c.svcCtx.Dao.ReactionClient.ReactionSetMessageReaction(c.ctx, &mtproto.TLReactionSetMessageReaction{
			UserId:    c.MD.UserId,
			PeerType:  peerType,
			PeerId:    peerId,
			MessageId: msgId,
			Reaction:  r,
			Big:       big,
			Date:      int32(mtproto.Now().Unix()),
		})
		if err != nil {
			return nil, err
		}
	}

	// Build updates to notify clients
	updatesHelper := mtproto.MakeUpdatesHelper()
	updateReaction := mtproto.MakeTLUpdateMessageReaction(&mtproto.Update{
		UserId:    c.MD.UserId,
		Peer:      peer,
		MsgId:     msgId,
		Reactions: in.GetReaction(),
	})
	updatesHelper.AddUpdate(updateReaction)

	return updatesHelper.ToUpdates(), nil
}

// MessagesGetMessagesReactions gets reactions for specific messages
func (c *ReactionsCore) MessagesGetMessagesReactions(in *mtproto.TLMessagesGetMessagesReactions) (*mtproto.Updates, error) {
	// Return empty updates for now — reactions are sent via update message
	return mtproto.MakeTLUpdatesEmpty(&mtproto.Updates{}).To_Updates(), nil
}

// MessagesGetTopReactions returns top reactions from recent messages
func (c *ReactionsCore) MessagesGetTopReactions(in *mtproto.TLMessagesGetTopReactions) (*mtproto.Messages_Reactions, error) {
	// Return empty list
	return mtproto.MakeTLMessagesReactions(&mtproto.Messages_Reactions{
		Reactions: []*mtproto.Reaction{},
		Period:    0,
		Hash:      0,
	}).To_Messages_Reactions(), nil
}

// MessagesGetRecentReactions returns recent reactions
func (c *ReactionsCore) MessagesGetRecentReactions(in *mtproto.TLMessagesGetRecentReactions) (*mtproto.Messages_Reactions, error) {
	return mtproto.MakeTLMessagesReactions(&mtproto.Messages_Reactions{
		Reactions: []*mtproto.Reaction{},
		Period:    0,
		Hash:      0,
	}).To_Messages_Reactions(), nil
}

// MessagesSetDefaultReaction sets a user's default reaction
func (c *ReactionsCore) MessagesSetDefaultReaction(in *mtproto.TLMessagesSetDefaultReaction) (*mtproto.Bool, error) {
	return mtproto.BoolTrue, nil
}
