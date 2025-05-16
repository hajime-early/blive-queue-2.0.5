package main

import (
	"fmt"
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var (
	LineBoxItem   = make(map[string]*fyne.Container)
	batteryLabels = make(map[string]*widget.Label) // 新增电池标签映射
	mu            sync.Mutex
	paused        bool
)

func MakeCtrlUI() *fyne.Container {
	// 修改1：将vbox包裹在滚动容器中，并设置初始尺寸
	vbox := container.NewVBox()
	scrollContainer := container.NewVScroll(vbox)
	scrollContainer.SetMinSize(fyne.NewSize(600, 800))

	// 修改2：修正清空按钮的重要性设置方式（字段赋值代替方法调用）
	clearBtn := widget.NewButton("清空列表", func() {
		mu.Lock()
		// 直接清空所有队列数据和索引
		line.GiftLine = []GiftLine{}
		line.CommonLine = []Line{}
		line.GiftIndex = map[string]int{}
		line.GiftIndex = map[string]int{}
		SetLine(line)
		mu.Unlock()

		vbox.RemoveAll()
		// 清空界面元素映射
		for k := range LineBoxItem {
			delete(LineBoxItem, k)
		}
	})

	clearBtn.Importance = widget.DangerImportance // 正确字段赋值

	// 修改2：保持暂停按钮的一致性设置方式（原代码正确无需修改）
	bottomButtons := container.NewHBox(
		layout.NewSpacer(),
		clearBtn, // 使用修改后的按钮变量
		func() *widget.Button {
			pauseButton := widget.NewButton("暂停排队", nil)
			pauseButton.Importance = widget.WarningImportance // 初始颜色为红色
			pauseButton.OnTapped = func() {
				mu.Lock()
				paused = !paused
				if paused {
					pauseButton.SetText("恢复排队")
					pauseButton.Importance = widget.SuccessImportance // 暂停时改为绿色
				} else {
					pauseButton.SetText("暂停排队")
					pauseButton.Importance = widget.WarningImportance // 恢复时改回红色
				}
				pauseButton.Refresh() // 刷新按钮样式
				mu.Unlock()
			}
			return pauseButton
		}(),
		layout.NewSpacer(),
	)

	mainContainer := container.NewBorder(nil, bottomButtons, nil, nil, scrollContainer)
	var (
		GiftLength   int
		CommonLength int
	)
	go func() {
		for {
			// 修改点：将OldLine获取移入循环内部，确保每次获取最新数据
			OldLine := line  // 移动这行到循环内部
			// 修改后条件判断将正确触发初始数据加载
			if len(OldLine.GiftLine) != GiftLength || len(OldLine.CommonLine) != CommonLength {
				fyne.Do(func() { // 新增：将整个UI更新操作包裹在fyne.Do中
					vbox.RemoveAll()
				})
				for idx, i2 := range OldLine.GiftLine {
					LineTemp := i2
					mu.Lock()
					// 新增礼物图标
					giftIcon := canvas.NewText("🎁", color.White) // 使用礼物符号
					giftIcon.SetMinSize(fyne.NewSize(16, 16))
					// 添加带序号的文本
					indexText := canvas.NewText(fmt.Sprintf("%d.", idx+1), color.White)
					// 礼物队列状态按钮初始化
					stateBtn := widget.NewButton("", nil)
					if LineTemp.IsOnline {
						stateBtn.SetText("离场")
						stateBtn.Importance = widget.MediumImportance // 新增：离场状态红色
					} else {
						stateBtn.SetText("在场")
						stateBtn.Importance = widget.HighImportance // 新增：在场状态绿色
					}

					// 创建文本对象并保留引用
					userNamePart := canvas.NewText(LineTemp.UserName, color.RGBA{255, 200, 0, 255}) // 金色偏黄 RGB(255,200,0)
					// 新增电池数量显示（单独一行）
					batteryLabel := widget.NewLabel(fmt.Sprintf("              礼物名：“%v”   累计电池：“%.2f” ", LineTemp.GiftName, LineTemp.GiftPrice))
					batteryLabels[LineTemp.OpenID] = batteryLabel // 记录标签引用
					statusSuffix := canvas.NewText("", color.White)
					if !LineTemp.IsOnline {
						statusSuffix.Text = "（不在）"
					}

					// 礼物队列按钮点击事件
					stateBtn.OnTapped = func() {
						mu.Lock()
						LineTemp.IsOnline = !LineTemp.IsOnline
						if LineTemp.IsOnline {
							stateBtn.SetText("离场")
							stateBtn.Importance = widget.MediumImportance // 新增：更新颜色
							statusSuffix.Text = ""
						} else {
							stateBtn.SetText("在场")
							stateBtn.Importance = widget.HighImportance // 新增：更新颜色
							statusSuffix.Text = "（不在）"
						}

						UpdateUserStatus(LineTemp.OpenID, LineTemp.IsOnline)
						statusSuffix.Refresh()
						stateBtn.Refresh()
						mu.Unlock()
					}

					LineBoxItem[LineTemp.OpenID] = container.NewHBox(
						container.NewVBox(
							container.NewHBox(
								giftIcon,
								indexText,
								userNamePart,
								statusSuffix,
							),
							container.NewHBox(
								container.NewCenter(batteryLabel), // 使用动态Label
							),
						),
						layout.NewSpacer(),
						stateBtn,
						widget.NewButton("删除", func() {
							fyne.Do(func() {
								vbox.Remove(LineBoxItem[LineTemp.OpenID])
							})
							DeleteLine(LineTemp.OpenID)
							mu.Lock()
							delete(LineBoxItem, LineTemp.OpenID)
							delete(batteryLabels, LineTemp.OpenID) // 删除时同步清理标签引用
							mu.Unlock()

							CommonLength = len(OldLine.GiftLine)
						}))
					mu.Unlock()
					fyne.Do(func() { // 新增：包裹添加操作
						vbox.Add(LineBoxItem[LineTemp.OpenID])
					})
				}

				if len(OldLine.CommonLine) != 0 {
					for cIdx, i2 := range OldLine.CommonLine {
						LineTemp := i2
						mu.Lock()
						chatIcon := canvas.NewText("💬", color.White) // 使用对话气泡符号
						chatIcon.SetMinSize(fyne.NewSize(16, 16))
						// 添加带独立序号的文本（从礼物队列长度+1开始）
						commonIndexText := canvas.NewText(fmt.Sprintf("%d.", len(OldLine.GiftLine)+cIdx+1), color.White)
						// 普通队列状态按钮初始化
						commonStateBtn := widget.NewButton("", nil)
						if LineTemp.IsOnline {
							commonStateBtn.SetText("离场")
							commonStateBtn.Importance = widget.MediumImportance // 新增：离场状态红色
						} else {
							commonStateBtn.SetText("在场")
							commonStateBtn.Importance = widget.HighImportance // 新增：在场状态绿色
						}

						commonNamePart := canvas.NewText(LineTemp.UserName, color.RGBA{135, 206, 235, 255}) // 天蓝色 RGB(135,206,235)
						commonStatusSuffix := canvas.NewText("", color.White)                               // 新增状态后缀文本（白色）
						if !LineTemp.IsOnline {
							commonStatusSuffix.Text = "（不在）"
						}

						// 普通队列按钮点击事件
						commonStateBtn.OnTapped = func() {
							mu.Lock()
							LineTemp.IsOnline = !LineTemp.IsOnline
							if LineTemp.IsOnline {
								commonStateBtn.SetText("离场")
								commonStateBtn.Importance = widget.MediumImportance // 新增：更新颜色
								commonStatusSuffix.Text = ""
							} else {
								commonStateBtn.SetText("在场")
								commonStateBtn.Importance = widget.HighImportance // 新增：更新颜色
								commonStatusSuffix.Text = "（不在）"
							}

							UpdateUserStatus(LineTemp.OpenID, LineTemp.IsOnline)
							commonStatusSuffix.Refresh()
							commonStateBtn.Refresh()
							mu.Unlock()
						}

						LineBoxItem[LineTemp.OpenID] = container.NewHBox(
							container.NewHBox(
								chatIcon,        // 新增普通队列图标
								commonIndexText, // 新增独立序号文本
								commonNamePart,
								commonStatusSuffix, // 保持普通队列状态后缀位置
							),
							layout.NewSpacer(),
							commonStateBtn,
							widget.NewButton("删除", func() {
								fyne.Do(func() { // 修复：包裹普通队列删除操作
									vbox.Remove(LineBoxItem[LineTemp.OpenID])
								})
								DeleteLine(LineTemp.OpenID)
								mu.Lock()
								delete(LineBoxItem, LineTemp.OpenID)
								mu.Unlock()
								CommonLength = len(OldLine.CommonLine)
							}),
						)
						mu.Unlock()
						fyne.Do(func() { // 新增：包裹添加操作
							vbox.Add(LineBoxItem[LineTemp.OpenID])
						})
					}
				}

				fyne.Do(func() { // 新增：包裹最后刷新操作
					vbox.Refresh()
				})

				// 修改点：在数据更新后重新获取最新数据
				GiftLength = len(OldLine.GiftLine)
				CommonLength = len(OldLine.CommonLine)
			}

			// 新增电池数据实时更新逻辑
			// 修改点：使用最新line数据替代OldLine
			for openID, label := range batteryLabels {
				for _, gift := range line.GiftLine {  // 修改为使用当前line数据
					if gift.OpenID == openID {
						newText := fmt.Sprintf("              礼物名：“%v”   累计电池：“%.2f” ", gift.GiftName, gift.GiftPrice)
						if label.Text != newText {
							fyne.Do(func() {
								label.SetText(newText)
							})
						}
						break
					}
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()
	return mainContainer
}
