// Stickers RPC handlers — implementation
package core

import (
	"context"
	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/bff/stickers/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/teamgram/marmota/pkg/metadata"
)

type StickersCore struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	MD *metadata.RpcMetadata
}

func New(ctx context.Context, svcCtx *svc.ServiceContext) *StickersCore {
	return &StickersCore{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		MD:     metadata.RpcMetadataFromIncoming(ctx),
	}
}

// MessagesGetAllStickers returns all installed sticker sets
func (c *StickersCore) MessagesGetAllStickers(in *mtproto.TLMessagesGetAllStickers) (*mtproto.Messages_AllStickers, error) {
	sets, err := c.svcCtx.Dao.StickerClient.StickerGetAllStickerSets(c.ctx, &mtproto.TLStickerGetAllStickerSets{
		UserId: c.MD.UserId,
	})
	if err != nil {
		return mtproto.MakeTLMessagesAllStickers(&mtproto.Messages_AllStickers{
			Hash: 0,
			Sets: []*mtproto.StickerSet{},
		}).To_Messages_AllStickers(), nil
	}
	return sets, nil
}

// MessagesGetFavedStickers returns favorite stickers
func (c *StickersCore) MessagesGetFavedStickers(in *mtproto.TLMessagesGetFavedStickers) (*mtproto.Messages_FavedStickers, error) {
	return mtproto.MakeTLMessagesFavedStickers(&mtproto.Messages_FavedStickers{
		Hash: 0,
		Packs: []*mtproto.StickerPack{},
		Stickers: []*mtproto.Document{},
	}).To_Messages_FavedStickers(), nil
}

// MessagesGetRecentStickers returns recent stickers
func (c *StickersCore) MessagesGetRecentStickers(in *mtproto.TLMessagesGetRecentStickers) (*mtproto.Messages_RecentStickers, error) {
	return mtproto.MakeTLMessagesRecentStickers(&mtproto.Messages_RecentStickers{
		Hash: 0,
		Packs: []*mtproto.StickerPack{},
		Stickers: []*mtproto.Document{},
		Dates: []int64{},
	}).To_Messages_RecentStickers(), nil
}

// MessagesGetFeaturedStickers returns featured sticker sets
func (c *StickersCore) MessagesGetFeaturedStickers(in *mtproto.TLMessagesGetFeaturedStickers) (*mtproto.Messages_FeaturedStickers, error) {
	return mtproto.MakeTLMessagesFeaturedStickers(&mtproto.Messages_FeaturedStickers{
		Hash: 0,
		Sets: []*mtproto.StickerSetCovered{},
		Unread: []int64{},
	}).To_Messages_FeaturedStickers(), nil
}

// MessagesGetArchivedStickers returns archived sticker sets
func (c *StickersCore) MessagesGetArchivedStickers(in *mtproto.TLMessagesGetArchivedStickers) (*mtproto.Messages_ArchivedStickers, error) {
	return mtproto.MakeTLMessagesArchivedStickers(&mtproto.Messages_ArchivedStickers{
		Count: 0,
		Sets: []*mtproto.StickerSetCovered{},
	}).To_Messages_ArchivedStickers(), nil
}

// MessagesGetStickers returns stickers matching an emoticon
func (c *StickersCore) MessagesGetStickers(in *mtproto.TLMessagesGetStickers) (*mtproto.Messages_Stickers, error) {
	return mtproto.MakeTLMessagesStickers(&mtproto.Messages_Stickers{
		Hash:     0,
		Stickers: []*mtproto.Document{},
	}).To_Messages_Stickers(), nil
}

// MessagesGetStickerSet returns a specific sticker set
func (c *StickersCore) MessagesGetStickerSet(in *mtproto.TLMessagesGetStickerSet) (*mtproto.Messages_StickerSet, error) {
	set, err := c.svcCtx.Dao.StickerClient.StickerGetStickerSet(c.ctx, &mtproto.TLStickerGetStickerSet{
		StickerSet: in.GetStickerset(),
		Hash:       in.GetHash(),
	})
	if err != nil {
		return mtproto.MakeTLMessagesStickerSetNotModified(&mtproto.Messages_StickerSet{}).To_Messages_StickerSet(), nil
	}
	return set, nil
}

// MessagesGetMaskStickers returns mask stickers
func (c *StickersCore) MessagesGetMaskStickers(in *mtproto.TLMessagesGetMaskStickers) (*mtproto.Messages_AllStickers, error) {
	return mtproto.MakeTLMessagesAllStickers(&mtproto.Messages_AllStickers{
		Hash: 0,
		Sets: []*mtproto.StickerSet{},
	}).To_Messages_AllStickers(), nil
}

// MessagesGetEmojiStickers returns emoji stickers
func (c *StickersCore) MessagesGetEmojiStickers(in *mtproto.TLMessagesGetEmojiStickers) (*mtproto.Messages_AllStickers, error) {
	return mtproto.MakeTLMessagesAllStickers(&mtproto.Messages_AllStickers{
		Hash: 0,
		Sets: []*mtproto.StickerSet{},
	}).To_Messages_AllStickers(), nil
}

// MessagesInstallStickerSet installs a sticker set for the user
func (c *StickersCore) MessagesInstallStickerSet(in *mtproto.TLMessagesInstallStickerSet) (*mtproto.Messages_StickerSetInstallResult, error) {
	result, err := c.svcCtx.Dao.StickerClient.StickerInstallStickerSet(c.ctx, &mtproto.TLStickerInstallStickerSet{
		UserId:     c.MD.UserId,
		StickerSet: in.GetStickerset(),
		Archived:   false,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// MessagesUninstallStickerSet uninstalls a sticker set
func (c *StickersCore) MessagesUninstallStickerSet(in *mtproto.TLMessagesUninstallStickerSet) (*mtproto.Bool, error) {
	_, err := c.svcCtx.Dao.StickerClient.StickerUninstallStickerSet(c.ctx, &mtproto.TLStickerUninstallStickerSet{
		UserId:     c.MD.UserId,
		StickerSet: in.GetStickerset(),
	})
	if err != nil {
		return mtproto.BoolFalse, nil
	}
	return mtproto.BoolTrue, nil
}

// MessagesReorderStickerSets reorders installed sticker sets
func (c *StickersCore) MessagesReorderStickerSets(in *mtproto.TLMessagesReorderStickerSets) (*mtproto.Bool, error) {
	_, err := c.svcCtx.Dao.StickerClient.StickerReorderStickerSets(c.ctx, &mtproto.TLStickerReorderStickerSets{
		UserId: c.MD.UserId,
		Order:  in.GetOrder(),
	})
	if err != nil {
		return mtproto.BoolFalse, nil
	}
	return mtproto.BoolTrue, nil
}

// MessagesSearchStickerSets searches for sticker sets
func (c *StickersCore) MessagesSearchStickerSets(in *mtproto.TLMessagesSearchStickerSets) (*mtproto.Messages_FoundStickerSets, error) {
	return mtproto.MakeTLMessagesFoundStickerSets(&mtproto.Messages_FoundStickerSets{
		Hash: 0,
		Sets: []*mtproto.StickerSetCovered{},
	}).To_Messages_FoundStickerSets(), nil
}
