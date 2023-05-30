package main

import (
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"github.com/skip2/go-qrcode"
	"wechatbot/config"
	"wechatbot/robot"
)

// ConsoleQrCode 控制台打印二维码
func ConsoleQrCode(uuid string) {
	q, _ := qrcode.New("https://login.weixin.qq.com/l/"+uuid, qrcode.Low)
	fmt.Println(q.ToString(true))
}

// LoginCallBack 登录回调
func LoginCallBack(body openwechat.CheckLoginResponse) {
	fmt.Println(string(body))
}

// dispatcher 消息分发器
var dispatcher = openwechat.NewMessageMatchDispatcher()

// MessageTextHandler 只处理消息类型为文本类型的消息
func MessageTextHandler(ctx *openwechat.MessageContext) {
	msg := ctx.Message
	robot.TextReplay(*msg)
}

// MessageImageHandler 只处理消息类型为图片类型的消息
func MessageImageHandler(ctx *openwechat.MessageContext) {
	msg := ctx.Message
	robot.ImgReplay(*msg)

}

// MessageVoiceHandler 只处理消息类型为图片类型的消息
func MessageVoiceHandler(ctx *openwechat.MessageContext) {
	msg := ctx.Message
	msg.ReplyText("声音")
}

// MessageFriendAddHandler 只处理消息类型为图片类型的消息
func MessageFriendAddHandler(ctx *openwechat.MessageContext) {
	msg := ctx.Message
	robot.FriendApply(*msg)
}

func init() {
	dispatcher.OnText(MessageTextHandler)
	dispatcher.OnImage(MessageImageHandler)
	dispatcher.OnVoice(MessageVoiceHandler)
	dispatcher.OnFriendAdd(MessageFriendAddHandler)
}

func main() {
	bot := openwechat.DefaultBot(openwechat.Desktop) // 桌面模式
	// 注册登陆二维码回调
	bot.UUIDCallback = ConsoleQrCode
	// 登录成功回调
	bot.LoginCallBack = LoginCallBack
	// 注册消息处理函数
	bot.MessageHandler = dispatcher.AsMessageHandler()

	// 创建热存储容器
	reloadStorage := openwechat.NewFileHotReloadStorage("storage.json")
	defer reloadStorage.Close()
	// 登陆
	if err := bot.PushLogin(reloadStorage, openwechat.NewRetryLoginOption()); err != nil {
		fmt.Println(err)
		return
	}

	// 获取登陆的用户
	self, err := bot.GetCurrentUser()
	config.RobotFillCallback(self.User.NickName)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 获取所有的好友
	friends, err := self.Friends()
	fmt.Println(friends, err)

	// 获取所有的群组
	groups, err := self.Groups()
	fmt.Println(groups, err)

	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	bot.Block()
}
