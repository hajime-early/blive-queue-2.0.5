package main

import (
	"image/color"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func MakeConfigUI(Windows fyne.Window, Config RunConfig) *fyne.Container {
	Windows.SetTitle("配置页面")

	IdCodeInput := widget.NewEntry()
	IdCodeInput.Text = Config.IdCode
	IdCodeInput.SetPlaceHolder("个人身份码")
	OpenFanfan := widget.NewButton("打开饭饭获取身份码", func() {
		err := AgreeOpenUrl("https://play-live.bilibili.com/")
		if err != nil {
			return
		}
	})

	LineKeyInput := widget.NewEntry()
	if Config.LineKey == "" {
		LineKeyInput.Text = "排队"
	} else {
		LineKeyInput.Text = Config.LineKey
	}
	LineKeyInput.SetPlaceHolder("请输入排队关键词")

	GiftJoinLine := widget.NewCheck("当有用户赠送大于设定值的礼物时自动加入队列", func(b bool) {})
	GiftJoinLine.Checked = Config.AutoJoinGiftLine

	GiftPriceDisplaySwitch := widget.NewCheck("是否显示礼物价格", func(b bool) {})
	GiftPriceDisplaySwitch.Checked = Config.GiftPriceDisplay

	IsOnlyGiftSwitch := widget.NewCheck("是否开启   <!->仅限<-!>   付费用户排队(舰长/礼物)", func(status bool) {
	})

	IsOnlyGiftSwitch.OnChanged = func(status bool) {
		if status {
			dialog.ShowConfirm("警告", "开启后只有舰长和送礼物的用户才能加入队列", func(b bool) {
				IsOnlyGiftSwitch.SetChecked(status)
			}, Windows)
		}
	}

	IsOnlyGiftSwitch.Checked = Config.IsOnlyGift

	Guard := canvas.NewText("舰长", color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if !Config.GuardPrintColor.IsEmpty() {
		Guard.Color = Config.GuardPrintColor.ToRGBA()
	}

	Gift := canvas.NewText("礼物用户", color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if !Config.GuardPrintColor.IsEmpty() {
		Gift.Color = Config.GiftPrintColor.ToRGBA()
	}

	Normal := canvas.NewText("普通用户", color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if !Config.CommonPrintColor.IsEmpty() {
		Normal.Color = Config.CommonPrintColor.ToRGBA()
	}

	DmDisplayColor := canvas.NewText("弹幕", color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if !Config.DmDisplayColor.IsEmpty() {
		DmDisplayColor.Color = Config.DmDisplayColor.ToRGBA()
	}
	TransparentBackgroundCheck := widget.NewCheck("开启排队展示无背景色 UI", func(b bool) {
	})
	TransparentBackgroundCheck.Checked = Config.TransparentBackground
	SelectLineColor := container.NewVBox(
		widget.NewLabel("请选择队列显示颜色\n当然，您可以在配置文件中自定义"),
		Guard,
		MakeSelectColor(Guard),
		Gift,
		MakeSelectColor(Gift),
		Normal,
		MakeSelectColor(Normal),
	)

	GiftPriceInput := widget.NewEntry()
	GiftPriceInput.SetPlaceHolder("加入队列的礼物价格门槛(电池)")
	if Config.GiftLinePrice > 0 {
		GiftPriceInput.Text = strconv.FormatFloat(Config.GiftLinePrice, 'f', -1, 64)
	}

	DisplayQueSize := widget.NewCheck("显示当前队列长度", func(b bool) {})
	DisplayQueSize.Checked = Config.CurrentQueueSizeDisplay

	EnableMusicServer := widget.NewCheck("启用音乐服务器", func(b bool) {})
	EnableMusicServer.Checked = Config.EnableMusicServer

	EnableDmDisplayNoSleep := widget.NewCheck("弹幕页面显示不休眠(移动端实验性)", func(b bool) {})
	EnableDmDisplayNoSleep.Checked = Config.DmDisplayNoSleep

	AutoScrollLine := widget.NewCheck("队列自动滚动展示", func(b bool) {})
	AutoScrollLine.Checked = Config.AutoScrollLine

	//滚动间隔
	ScrollIntervalInput := widget.NewEntry()
	ScrollIntervalInput.SetPlaceHolder("滚动间隔(秒)")
	if Config.ScrollInterval > 0 {
		ScrollIntervalInput.Text = strconv.Itoa(Config.ScrollInterval / 2)
	}

	LineMaxLengthInput := widget.NewEntry()
	LineMaxLengthInput.SetPlaceHolder("队列最大容量")
	if Config.MaxLineCount > 0 {
		LineMaxLengthInput.Text = strconv.Itoa(Config.MaxLineCount)
	}

	StartButton := widget.NewButton("保存配置并开始", func() {
		GiftLinePriceFloat64, err := strconv.ParseFloat(GiftPriceInput.Text, 10)
		LineMaxLengthInt, err := strconv.Atoi(LineMaxLengthInput.Text)
		ScrollIntervalInt, err := strconv.Atoi(ScrollIntervalInput.Text)

		switch {
		case len(IdCodeInput.Text) == 0:
			dialog.ShowError(DisplayError{Message: "房间号不能为空"}, Windows)
			return
		case GiftJoinLine.Checked && GiftLinePriceFloat64 <= 0:
			dialog.ShowError(DisplayError{Message: "礼物价格应该大于0"}, Windows)
			return

		case LineMaxLengthInt <= 0:
			dialog.ShowError(DisplayError{Message: "队列最大容量应该大于0"}, Windows)
			return
		}

		if LineKeyInput.Text == "" {
			LineKeyInput.Text = "排队"
		}

		SaveConfig := RunConfig{
			IdCode:                  IdCodeInput.Text,
			GuardPrintColor:         ToLineColor(Guard.Color),
			GiftPriceDisplay:        GiftPriceDisplaySwitch.Checked,
			GiftPrintColor:          ToLineColor(Gift.Color),
			GiftLinePrice:           GiftLinePriceFloat64,
			CommonPrintColor:        ToLineColor(Normal.Color),
			DmDisplayColor:          ToLineColor(DmDisplayColor.Color),
			LineKey:                 LineKeyInput.Text,
			IsOnlyGift:              IsOnlyGiftSwitch.Checked,
			AutoJoinGiftLine:        GiftJoinLine.Checked,
			TransparentBackground:   TransparentBackgroundCheck.Checked,
			MaxLineCount:            LineMaxLengthInt,
			CurrentQueueSizeDisplay: DisplayQueSize.Checked,
			EnableMusicServer:       EnableMusicServer.Checked,
			DmDisplayNoSleep:        EnableDmDisplayNoSleep.Checked,
			ScrollInterval:          ScrollIntervalInt * 2,
			AutoScrollLine:          AutoScrollLine.Checked,
		}

		KeyWordMatchMap = make(map[string]bool)
		KeyWordMatchInit(SaveConfig.LineKey)

		if err != nil {
			dialog.ShowError(err, Windows)
		} else {
			globalConfiguration = SaveConfig
			SetConfig(SaveConfig)
			dialog.ShowInformation("保存成功", "配置已保存,如果涉及身份码修改,请重启", Windows)
			Restart()
			time.Sleep(1 * time.Second)
			Windows.SetContent(MakeMainUI(Windows, SaveConfig))

		}
	})
	return container.NewVBox(
		IdCodeInput,
		OpenFanfan,
		LineKeyInput,
		IsOnlyGiftSwitch,
		GiftPriceDisplaySwitch,
		TransparentBackgroundCheck,
		SelectLineColor,
		GiftJoinLine,
		GiftPriceInput,
		DisplayQueSize,
		EnableMusicServer,
		EnableDmDisplayNoSleep,
		LineMaxLengthInput,
		AutoScrollLine,
		ScrollIntervalInput,

		StartButton,
	)
}

func MakeSelectColor(text *canvas.Text) *fyne.Container {
	return container.NewHBox(
		widget.NewButton("暗蓝", func() {
			text.Color = color.RGBA{R: 6, G: 68, B: 255, A: 255}
			text.Refresh()
		}),
		widget.NewButton("深绿", func() {
			text.Color = color.RGBA{R: 18, G: 146, B: 14, A: 255}
			text.Refresh()
		}),
		widget.NewButton("淡蓝", func() {
			text.Color = color.RGBA{R: 58, G: 150, B: 221, A: 255}
			text.Refresh()
		}),
		widget.NewButton("红色", func() {
			text.Color = color.RGBA{R: 255, G: 26, B: 45, A: 255}
			text.Refresh()
		}),
		widget.NewButton("暗紫", func() {
			text.Color = color.RGBA{R: 187, G: 31, B: 211, A: 255}
			text.Refresh()
		}),
		widget.NewButton("暗棕", func() {
			text.Color = color.RGBA{R: 193, G: 156, B: 0, A: 255}
			text.Refresh()
		}),
		widget.NewButton("蓝色", func() {
			text.Color = color.RGBA{R: 59, G: 120, B: 255, A: 255}
			text.Refresh()
		}),
		widget.NewButton("绿色", func() {
			text.Color = color.RGBA{R: 22, G: 198, B: 12, A: 255}
			text.Refresh()
		}),
		widget.NewButton("亮蓝", func() {
			text.Color = color.RGBA{R: 100, G: 221, B: 221, A: 255}
			text.Refresh()
		}),
		widget.NewButton("大红", func() {
			text.Color = color.RGBA{R: 231, G: 72, B: 86, A: 255}
			text.Refresh()
		}),
		widget.NewButton("紫色", func() {
			text.Color = color.RGBA{R: 180, G: 0, B: 158, A: 255}
			text.Refresh()
		}),
		widget.NewButton("黄色", func() {
			text.Color = color.RGBA{R: 249, G: 241, B: 165, A: 255}
			text.Refresh()
		}),
		widget.NewButton("自定义选择", func() {
			MakeColorPicker(text)
		}),
	)
}

func MakeColorPicker(text *canvas.Text) {
	ColorPicker := dialog.NewColorPicker("颜色选择", "", func(c color.Color) {
		text.Color = c
		text.Refresh()
	}, MainWindows)
	ColorPicker.Advanced = true
	ColorPicker.Show()
}
