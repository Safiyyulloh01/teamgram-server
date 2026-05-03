// Bots RPC handlers — implementation over existing service layer
package core

import (
	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/bff/bots/internal/svc"
	"github.com/teamgram/teamgram-server/app/service/biz/user/user"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/teamgram/marmota/pkg/metadata"
	"context"
)

type BotsCore struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	MD *metadata.RpcMetadata
}

func New(ctx context.Context, svcCtx *svc.ServiceContext) *BotsCore {
	return &BotsCore{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		MD:     metadata.RpcMetadataFromIncoming(ctx),
	}
}

// AuthImportBotAuthorization imports bot authorization
func (c *BotsCore) AuthImportBotAuthorization(in *mtproto.TLAuthImportBotAuthorization) (*mtproto.Auth_Authorization, error) {
	apiId := in.GetApiId()
	apiHash := in.GetApiHash()
	token := in.GetToken()

	// Look up bot by token
	userData, err := c.svcCtx.Dao.UserClient.UserGetUserDataByToken(c.ctx, &user.TLUserGetUserDataByToken{
		Token: token,
	})
	if err != nil {
		return nil, mtproto.ErrAuthTokenInvalid
	}

	// Verify API credentials
	appInfo, err := c.svcCtx.Dao.AuthsessionClient.AuthsessionCheckApiId(c.ctx, &mtproto.TLAuthsessionCheckApiId{
		ApiId: apiId,
	})
	if err != nil || !mtproto.FromBool(appInfo) {
		return nil, mtproto.ErrApiIdInvalid
	}

	// Create auth session for bot
	session, err := c.svcCtx.Dao.AuthsessionClient.AuthsessionCreateAuthSession(c.ctx, &mtproto.TLAuthsessionCreateAuthSession{
		UserId:     userData.GetBotId(),
		ApiId:      apiId,
		DcId:       1,
		DeviceModel: "bot",
	})
	if err != nil {
		return nil, err
	}

	// Return authorization with bot user
	botUser, _ := c.svcCtx.Dao.UserClient.UserGetMutableUsers(c.ctx, &user.TLUserGetMutableUsers{
		Id: []int64{userData.GetBotId()},
	})

	return mtproto.MakeTLAuthAuthorization(&mtproto.Auth_Authorization{
		User: botUser.GetUserListByIdList(c.MD.UserId, userData.GetBotId())[0],
	}).To_Auth_Authorization(), nil
}

// BotsSendCustomRequest sends a custom request to a bot
func (c *BotsCore) BotsSendCustomRequest(in *mtproto.TLBotsSendCustomRequest) (*mtproto.DataJSON, error) {
	// For now, return an empty response
	return mtproto.MakeTLDataJSON(&mtproto.DataJSON{
		Data: "{}",
	}).To_DataJSON(), nil
}

// BotsAnswerWebhookJSONQuery answers a webhook JSON query
func (c *BotsCore) BotsAnswerWebhookJSONQuery(in *mtproto.TLBotsAnswerWebhookJSONQuery) (*mtproto.Bool, error) {
	return mtproto.BoolTrue, nil
}

// BotsSetBotCommands sets bot commands for a specific scope
func (c *BotsCore) BotsSetBotCommands(in *mtproto.TLBotsSetBotCommands) (*mtproto.Bool, error) {
	_, err := c.svcCtx.Dao.UserClient.UserSetBotCommands(c.ctx, &user.TLUserSetBotCommands{
		UserId:   c.MD.UserId,
		BotId:    in.GetBot().GetUserId(),
		Scope:    in.GetScope(),
		Commands: in.GetCommands(),
	})
	if err != nil {
		return mtproto.BoolFalse, nil
	}
	return mtproto.BoolTrue, nil
}

// BotsResetBotCommands resets bot commands
func (c *BotsCore) BotsResetBotCommands(in *mtproto.TLBotsResetBotCommands) (*mtproto.Bool, error) {
	_, err := c.svcCtx.Dao.UserClient.UserResetBotCommands(c.ctx, &user.TLUserResetBotCommands{
		UserId: c.MD.UserId,
		BotId:  in.GetBot().GetUserId(),
		Scope:  in.GetScope(),
	})
	if err != nil {
		return mtproto.BoolFalse, nil
	}
	return mtproto.BoolTrue, nil
}

// BotsGetBotCommands retrieves bot commands
func (c *BotsCore) BotsGetBotCommands(in *mtproto.TLBotsGetBotCommands) (*mtproto.BotCommands, error) {
	commands, err := c.svcCtx.Dao.UserClient.UserGetBotCommands(c.ctx, &user.TLUserGetBotCommands{
		UserId: c.MD.UserId,
		BotId:  in.GetBot().GetUserId(),
		Scope:  in.GetScope(),
	})
	if err != nil {
		return mtproto.MakeTLBotCommands(&mtproto.BotCommands{
			Commands: []*mtproto.BotCommand{},
		}).To_BotCommands(), nil
	}
	return commands, nil
}

// BotsCheckBot checks if a bot exists
func (c *BotsCore) BotsCheckBot(in *mtproto.TLBotsCheckBot) (*mtproto.Bool, error) {
	isBot, err := c.svcCtx.Dao.UserClient.UserCheckBots(c.ctx, &user.TLUserCheckBots{
		UserId: c.MD.UserId,
		BotId:  in.GetBot().GetUserId(),
	})
	if err != nil {
		return mtproto.BoolFalse, nil
	}
	return isBot, nil
}

// MessagesGetBotCallbackAnswer gets the callback answer for an inline button
func (c *BotsCore) MessagesGetBotCallbackAnswer(in *mtproto.TLMessagesGetBotCallbackAnswer) (*mtproto.Messages_BotCallbackAnswer, error) {
	// Return empty callback answer
	return mtproto.MakeTLMessagesBotCallbackAnswer(&mtproto.Messages_BotCallbackAnswer{
		Alert:    false,
		HasUrl:   false,
		NativeUi: false,
		Message:  "",
		Url:      "",
		CacheTime: 0,
	}).To_Messages_BotCallbackAnswer(), nil
}

// MessagesSetBotPrecheckoutResults sets pre-checkout results
func (c *BotsCore) MessagesSetBotPrecheckoutResults(in *mtproto.TLMessagesSetBotPrecheckoutResults) (*mtproto.Bool, error) {
	return mtproto.BoolTrue, nil
}

// MessagesSetBotShippingResults sets shipping results
func (c *BotsCore) MessagesSetBotShippingResults(in *mtproto.TLMessagesSetBotShippingResults) (*mtproto.Bool, error) {
	return mtproto.BoolTrue, nil
}

// MessagesRequestAppWebView requests a miniapp webview
func (c *BotsCore) MessagesRequestAppWebView(in *mtproto.TLMessagesRequestAppWebView) (*mtproto.WebViewResult, error) {
	// Return an empty response — miniapp requires full enterprise implementation
	return mtproto.MakeTLWebViewResultUrl(&mtproto.WebViewResult{
		QueryId:  0,
		Url:      "",
		FullSize: false,
	}).To_WebViewResult(), nil
}
