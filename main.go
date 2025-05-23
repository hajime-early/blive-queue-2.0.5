package main

import (
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/vtb-link/bianka/basic"
	"github.com/vtb-link/bianka/live"
	"golang.org/x/exp/slog"
	"gopkg.in/natefinch/lumberjack.v2"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

//go:embed Resource/bilibili-line.svg
var icon []byte

var (
	App fyne.App

	RoomId                int
	MainWindows           fyne.Window
	CtrlWindows           fyne.Window
	SpecialUserSetWindows fyne.Window

	line                LineRow
	globalConfiguration RunConfig

	svgResource *fyne.StaticResource
)

var AppClient *live.Client
var GameId string
var CloseHeartbeatChan chan bool
var WsClient *basic.WsClient

var logger *slog.Logger

//var DanmuDataChan = make(chan *proto.CmdDanmuData, 20)

func main() {
	// 为全局变量赋值
	line.GiftIndex = make(map[string]int)
	line.CommonIndex = make(map[string]int)

	lineTemp, err := GetLine()
	if err == nil && !lineTemp.IsEmpty() {
		line = lineTemp
	}

	r := &lumberjack.Logger{
		Filename:   "./BLine.log",
		LocalTime:  true,
		MaxSize:    1,
		MaxAge:     3,
		MaxBackups: 5,
		Compress:   true,
	}

	logger = slog.New(slog.NewJSONHandler(r, nil))
	slog.SetDefault(logger)

	//go ResponseQueCtrl()

	// CleanOldVersion()

	svgResource = fyne.NewStaticResource("icon.svg", icon)
	// 资源初始化区域
	App = app.New()
	App.Settings().SetTheme(theme.DarkTheme())
	// 窗口大体定义区域

	//自改版，暂时用不到更新
	// NewVersion, UpdateStatus := CheckVersion()

	// if UpdateStatus {
	// 	UpdateWindows := App.NewWindow("有新版本")
	// 	UpdateWindows.Resize(fyne.NewSize(300, 300))
	// 	UpdateUI := MakeUpdateUI(NewVersion)
	// 	UpdateWindows.SetContent(UpdateUI)
	// 	UpdateWindows.Show()
	// }
	MainWindows = App.NewWindow("未初始化")
	MainWindows.SetIcon(svgResource)

	//var err error

	// 修改连接逻辑
	var retryCount int
	for {
		globalConfiguration, err = GetConfig()
		if err != nil {
			slog.Error("Get config Err", err)
			MainWindows.SetContent(MakeConfigUI(MainWindows, RunConfig{}))
			break
		}

		client, gameId, wsClient, closeChan := RoomConnect(globalConfiguration.IdCode)
		if client != nil { // 仅当连接成功时退出循环
			AppClient = client
			CloseHeartbeatChan = closeChan
			GameId = gameId
			WsClient = wsClient
			KeyWordMatchInit(globalConfiguration.LineKey)
			MainWindows.SetContent(MakeMainUI(MainWindows, globalConfiguration))
			break
		}

		// 连接失败时等待后重试
		retryCount++
		if retryCount > 3 { // 最大重试次数保护
			slog.Error("达到最大重试次数，停止连接")
			break
		}
		slog.Info(fmt.Sprintf("第%d次连接失败，5秒后重试...", retryCount))
		time.Sleep(5 * time.Second)
	}

	//初始化控制界面
	CtrlWindows = App.NewWindow("控制界面 点击两次 ╳ 退出")
	CtrlWindows.SetIcon(svgResource)
	// 关闭此窗口退出应用
	CtrlWindows.SetMaster()
	var ClickCount int
	CtrlWindows.SetCloseIntercept(func() {
		ClickCount++
		if ClickCount > 1 {
			CtrlWindows.Close()
			App.Quit()
			os.Exit(0)
		}
	})

	CtrlWindows.RequestFocus()

	CtrlWindows.Resize(fyne.NewSize(400, 600))
	CtrlUIContext := MakeCtrlUI()
	size := CtrlUIContext.Size()
	// 打印窗口尺寸
	fmt.Printf("Window width: %f, height: %f\n", size.Width, size.Height)
	CtrlWindows.SetContent(MakeCtrlUI())
	CtrlWindows.Show()

	go StartWebServer()
	MainWindows.Show()
	App.Run()
}
