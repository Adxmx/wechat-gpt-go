package robot

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"wechatbot/config"
	"wechatbot/openai"
)

// verifyToken 验证Token
var verifyToken string

// userDB 聊天上下文存储
var userDB map[string][]map[string]string

type Role struct {
	Name   string              `json:"name"`
	Prompt []map[string]string `json:"prompt"`
}

// roleDB 角色信息存储
var roleDB []Role

// conf 配置信息
var conf map[string]string

// GenerateToken 生成32位Token
func GenerateToken(length int) string {
	bytes := make([]byte, length/2)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(bytes)
}

// initUserDB 初始化用户上线文
func initUserDB(username string) {
	if _, flag := userDB[username]; !flag {
		userDB[username] = []map[string]string{}
	}
}

// updateUserDB 更新聊天上下文存储
func updateUserDB(username string, role string, content string) []map[string]string {
	userDB[username] = append(userDB[username], map[string]string{"role": role, "content": content})
	context := filterSlice(userDB[username], false)
	if len(context) > 4 {
		// 删除第一行记录
		context = context[1:]
	}
	userDB[username] = append(filterSlice(userDB[username], true), context...)
	return userDB[username]
}

// clearUserDB 清空聊天上下文存储
func clearUserDB(username string) {
	userDB[username] = filterSlice(userDB[username], true)
}

// 设置prompt
func setPromptDB(username string, roleId int) {
	userDB[username] = append(roleDB[roleId].Prompt, filterSlice(userDB[username], false)...)
}

// filterSlice 切片过滤
func filterSlice(userSlice []map[string]string, isPrompt bool) []map[string]string {
	var result []map[string]string
	filter := func(isPrompt bool, item map[string]string) bool {
		if isPrompt {
			return item["role"] == "system"
		} else {
			return item["role"] == "user" || item["role"] == "assistant"
		}
	}
	for _, item := range userSlice {
		if filter(isPrompt, item) {
			result = append(result, item)
		}
	}

	return result
}

// ImgReplay 图片消息回复
func ImgReplay(msg openwechat.Message) {
	resp, _ := msg.GetPicture()
	defer resp.Body.Close()
	path := fmt.Sprintf("/tmp/origin%s.jpg", GenerateToken(32))
	file, _ := os.Create(path)
	defer file.Close()
	io.Copy(file, resp.Body)
	// 图片填充正方形并转png
	path = square(path)
	// GPT变形，保存到本地
	path = imgDownload(openai.ImgVariate(path))
	// 返回
	img, _ := os.Open(path)
	msg.ReplyImage(img)
	msg.ReplyText("变形成功！")
}

// TextReplay 文本消息回复
func TextReplay(msg openwechat.Message) {
	// 获取用户信息
	user, _ := msg.Sender()
	// 判断是否群聊
	isGroup := user.IsGroup()
	// 获取用户标识(仅一次会话不变)
	username := user.UserName
	// 获取用户昵称(仅一次会话不变)
	nickname := user.NickName
	fmt.Println("=======文本发送======: ", username, nickname, ":", msg.Content)
	// 判断是否回复
	if !isReplay(*user, isGroup, msg.Content) {
		return
	}
	// 初始化用户信息
	initUserDB(username)
	// 定义返回的消息
	// 测试消息
	if msg.Content == "/adxm" {
		msg.ReplyText("adxm")
	} else {
		// 群聊 私聊
		if isGroup {
			content := msg.Content[strings.Index(msg.Content, " ")+1:]
			textAnalyzeAndReplay(content, username, *user, msg)
		} else {
			textAnalyzeAndReplay(msg.Content, username, *user, msg)
		}
	}
}

// isReplay 判断是否回复
func isReplay(user openwechat.User, isGroup bool, content string) bool {
	flag := false
	// 判断是否群聊
	if isGroup {
		// 判断是否@自己
		flag = strings.HasPrefix(content, "@"+config.Conf.RobotInfo.Nickname)
	} else {
		if user.IsFriend() {
			flag = true
		}
	}
	return flag
}

// TextAnalyze 消息分析
func textAnalyzeAndReplay(content string, username string, user openwechat.User, msg openwechat.Message) {
	if strings.HasPrefix(content, "/cmd ") {
		cmd := content[5:]
		replay, flag := cmdParse(cmd, username, user)
		if flag == "tip" {
			// 返回文本消息
			msg.ReplyText(replay)
		} else if flag == "img" {
			// 返回图片文件
			path := imgDownload(replay)
			img, _ := os.Open(path)
			msg.ReplyImage(img)
			msg.ReplyText("绘画完成！")
		}
	} else {
		// 获取用户上下文
		updateUserDB(username, "user", content)
		// TODO 请求GPT聊天接口
		replay := openai.ChatComplete(userDB[username])
		// 更新用户上下文
		updateUserDB(username, "assistant", replay)
		msg.ReplyText(replay)
	}
	fmt.Println(userDB[username])
}

// imgDownload 将图片下载保存到本地
func imgDownload(url string) string {
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	path := fmt.Sprintf("/tmp/%s.png", GenerateToken(32))
	file, _ := os.Create(path)
	defer file.Close()
	io.Copy(file, resp.Body)
	return path
}

// cmdParse 命令解析
func cmdParse(cmd string, username string, user openwechat.User) (string, string) {
	tip, flag := fmt.Sprintf("很抱歉，【%s】无法理解%s命令", config.Conf.RobotInfo.Nickname, cmd), "tip"
	if cmd == "help" {
		tip = fmt.Sprintf(`您好，【%s】为您服务！
1. 显示帮助信息
  /cmd help
2. 角色列表
  /cmd role
3. 扮演角色
  /cmd role NO.
4. AI绘图
  /cmd img DESC.
5. 清空聊天上下文
  /cmd clear`, config.Conf.RobotInfo.Nickname)
	} else if cmd == "clear" {
		tip = fmt.Sprintf("【%s】已清空上下文信息，可以开启新的话题了！", config.Conf.RobotInfo.Nickname)
		clearUserDB(username)
	} else if cmd == "role" {
		tip = fmt.Sprintf(`以下是角色列表:
%s
输入指令，【%s】将进入角色扮演模式。
例如：/cmd role 12 进入将进入猫娘角色扮演模式`, roleListText(), config.Conf.RobotInfo.Nickname)
	} else if matchResult, _ := regexp.MatchString("^role \\d+$", cmd); matchResult {
		roleId, err := strconv.Atoi(cmd[5:])
		if err != nil && roleId >= len(roleDB) {
			tip = fmt.Sprintf("【%s】没有%d角色信息，无法扮演！", config.Conf.RobotInfo.Nickname, roleId)
		} else {
			tip = fmt.Sprintf("【%s】已进入%s角色！", config.Conf.RobotInfo.Nickname, roleDB[roleId].Name)
			setPromptDB(username, roleId)
		}
	} else if matchResult, _ := regexp.MatchString("^img .+$", cmd); matchResult {
		prompt := cmd[4:]
		fmt.Println("======================prompt========================", prompt)
		tip, flag = openai.ImgCreate(prompt), "img"
	} else if cmd == "token" {
		tip = "很抱歉，您无权查看！"
		id := user.ID()
		fmt.Println(id)
		if admins := config.Conf.RobotInfo.Admin; len(admins) > 0 {
			for _, item := range admins {
				if id == item {
					tip = verifyToken
				}
			}
		} else {
			tip = verifyToken
		}
	}
	return tip, flag
}

// loadRole 加载json数据
func loadRole() {
	data, _ := ioutil.ReadFile("role.json")
	json.Unmarshal(data, &roleDB)
}

// roleListText 生成角色列表
func roleListText() string {
	var listText string = ""
	for index, role := range roleDB {
		listText += fmt.Sprintf("%d. %s\n", index, role.Name)
	}
	return listText
}

// FriendApply 好友添加
func FriendApply(msg openwechat.Message) {
	match := regexp.MustCompile(`content="(.*?)"`).FindStringSubmatch(msg.Content)
	user, _ := msg.Sender()
	fmt.Println("=======好友申请======", user.NickName, ":", match)
	if len(match) > 1 {
		if match[1] == verifyToken {
			friend, _ := msg.Agree()
			friend.SendText(config.Conf.RobotInfo.Prologue)
		}
	}
}

func init() {
	userDB = make(map[string][]map[string]string)
	// 加载json文件中角色信息
	loadRole()
	// 启动定时器
	go func() {
		verifyToken = GenerateToken(32)
		fmt.Println("=======验证Token======", verifyToken)
		for {
			<-time.NewTicker(1 * time.Hour).C
			verifyToken = GenerateToken(32)
			fmt.Println("=======验证Token======", verifyToken)
		}
	}()
}
