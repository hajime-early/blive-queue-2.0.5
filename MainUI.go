package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/atotto/clipboard"
	"golang.org/x/exp/slog"
)

//go:embed Resource/404.jpg
var Pic404 []byte

func MakeMainUI(Windows fyne.Window, Config RunConfig) *fyne.Container {
	Windows.SetTitle("主页面")
	var RoomInformationObtained RoomInfo
	for RoomId == 0 {
	}
	fmt.Println("主线程房间号", RoomId)
	RoomInformationObtained, err := GetRoomInfo(strconv.Itoa(RoomId))
	if err != nil {
		dialog.ShowError(DisplayError{Message: "获得房间信息错误 请重新输入房间号"}, Windows)
	}

	if RoomInformationObtained.Code != 0 {
		dialog.ShowError(DisplayError{Message: "房间号不存在，请检查是否输入正确"}, Windows)
	}
	// HACK:应付折扣礼物的临时功能函数
	GetRoomGiftData(RoomInformationObtained.Data.RoomId)

	TittleDisplay := container.NewHBox(
		canvas.NewText("标题:", color.White),
		canvas.NewText(RoomInformationObtained.Data.Title, color.White),
	)
	LiveStatusDisplay := container.NewHBox(
		canvas.NewText("当前状态:", color.White), []fyne.CanvasObject{
			canvas.NewText("未开播", color.White),
			canvas.NewText("直播中", color.RGBA{R: 100, G: 221, B: 221, A: 255}),
			canvas.NewText("轮播中", color.RGBA{R: 58, G: 150, B: 221, A: 255}),
		}[RoomInformationObtained.Data.LiveStatus])
	DescDisplay := container.NewHBox(
		canvas.NewText("描述:", color.White),
		widget.NewLabel(RemoveTags(RoomInformationObtained.Data.Description)),
	)

	var CoverDisplay io.Reader = bytes.NewReader(Pic404)
	get, err := http.Get(RoomInformationObtained.Data.UserCover)
	if err != nil {
		slog.Error("获取直播封面错误", err)
	} else {
		defer get.Body.Close()
		CoverDisplay = get.Body
	}

	LiveCoverDisplay := canvas.NewImageFromReader(CoverDisplay, "直播封面")
	LiveCoverDisplay.FillMode = canvas.ImageFillOriginal

	JumpToConfigUI := widget.NewButton("重新设置", func() {
		Windows.SetContent(MakeConfigUI(Windows, Config))
	})
	CopyLineUrlButton := widget.NewButton("复制排队组件Url", func() {
		err = clipboard.WriteAll("http://127.0.0.1:100/web")
		if err != nil {
			dialog.ShowError(DisplayError{"写入剪贴板错误"}, Windows)
			return
		}
	})
	CopyDmUrlButton := widget.NewButton("复制弹幕组件Url", func() {
		err := clipboard.WriteAll("http://127.0.0.1:100/dm")
		if err != nil {
			dialog.ShowError(DisplayError{"写入剪贴板错误"}, Windows)
			return
		}
	})

	CopyMusicUrlButton := widget.NewButton("复制音乐组件Url[仅在开启音乐插件后有效]", func() {
		err := clipboard.WriteAll("http://127.0.0.1:99/music")
		if err != nil {
			dialog.ShowError(DisplayError{"写入剪贴板错误"}, Windows)
			return
		}
	})

	// SpecialSetButton := widget.NewButton("特殊用户设置", func() {
	// 	//初始化特殊用户设置界面
	// 	SpecialUserSetWindows = App.NewWindow("特殊用户设置")
	// 	SpecialUserSetWindows.SetIcon(svgResource)
	// 	SpecialUserSetWindows.Resize(fyne.NewSize(400, 600))
	// 	SpecialUserSetWindows.SetContent(DisplaySpecialUserListUI())
	// 	SpecialUserSetWindows.Show()
	// })

	ReconnectButton := widget.NewButton("重连弹幕服务器", func() {
		Restart()
	})
	if !globalConfiguration.EnableMusicServer {
		CopyMusicUrlButton.Hide()
	}

	assist := container.NewHBox()
	if randomInt(0, 10) >= 6 {
		assist.Add(widget.NewButton("赞助作者", func() {
			assistDia := dialog.NewCustom("感谢您的赞助", "关闭", assistUI(), MainWindows)
			assistDia.Show()
			dialog.ShowError(DisplayError{Message: "赞助只会让我恰烂钱，并不会提供比别人更多的技术支持"}, MainWindows)
			dialog.ShowError(DisplayError{Message: "未成年人请勿赞助"}, MainWindows)
		}))
	}

	if RoomInformationObtained.Data.LiveStatus == 1 {
		LiveStarTimeDisplay := container.NewHBox(
			canvas.NewText("直播开始时间:", color.White),
			canvas.NewText(RoomInformationObtained.Data.LiveTime, color.White),
		)
		difference := CalculateTimeDifference(RoomInformationObtained.Data.LiveTime)

		LiveKeepTimeDisplay := container.NewHBox(
			canvas.NewText("已直播:", color.White),
			canvas.NewText(difference.String(), color.White),
		)

		return container.NewVBox(TittleDisplay, LiveStatusDisplay, DescDisplay, LiveCoverDisplay, LiveStarTimeDisplay, LiveKeepTimeDisplay, CopyLineUrlButton, CopyDmUrlButton, CopyMusicUrlButton, JumpToConfigUI, ReconnectButton, assist)
	} else {
		return container.NewVBox(TittleDisplay, LiveStatusDisplay, DescDisplay, LiveCoverDisplay, CopyLineUrlButton, CopyDmUrlButton, CopyMusicUrlButton, JumpToConfigUI, ReconnectButton, assist)
	}
}

func GetRoomInfo(RoomId string) (RoomInfo, error) {
	get, err := http.Get("https://api.live.bilibili.com/room/v1/Room/get_info?id=" + RoomId)
	if err != nil {
		return RoomInfo{}, err
	}
	all, err := io.ReadAll(get.Body)
	if err != nil {
		return RoomInfo{}, err
	}
	var GetRoomInfo RoomInfo
	err = json.Unmarshal(all, &GetRoomInfo)
	if err != nil {
		return RoomInfo{}, err
	}
	return GetRoomInfo, nil
}

func MakeSpecialManagerList(SpecialUserList map[string]int64) *fyne.Container {
	vbox := container.NewVBox()
	for k, v := range SpecialUserList {
		vbox.Add(container.NewHBox(canvas.NewText(k, color.White), canvas.NewText("到期时间:"+TimestampToTime(v).String(), color.White)))
	}
	return vbox
}

// TimestampToTime 时间戳转时间
func TimestampToTime(timestamp int64) (FormatTime time.Time) {
	return time.Unix(timestamp, 0)
}
