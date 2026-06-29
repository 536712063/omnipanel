// OmniPanel 主入口
//
// 编译:
//   cd ui && npm run build    (构建 Vue 3 前端)
//   wails build               (编译 Go 二进制 + 嵌入前端)
//
// 架构:
//   main.go → app.CreateWailsService() → Wails Application
//              ├── app.go       (前端绑定方法)
//              ├── agent/        (插件管理器)
//              ├── license/      (License 系统)
//              └── plugins/      (独立插件进程)
package main

import (
	"log"
	"os"

	"github.com/omnipanel/omnipanel/internal/app"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("OmniPanel 启动中...")

	// 创建应用核心实例
	panel := app.NewOmniPanel()

	// 配置 Wails 服务
	wailsApp := app.CreateWailsService(panel)

	// 创建主窗口
	mainWindow := wailsApp.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Title:  "OmniPanel - 全能面板",
		Width:  1400,
		Height: 900,
		MinWidth:  1000,
		MinHeight: 700,
		BackgroundColour: application.RGBA{R: 26, G: 26, B: 46, A: 255},
		URL: "/",
		Frameless: false,
		AlwaysOnTop: false,
		Fullscreen:  false,
		StartState:  application.WindowStateMaximised,
	})

	// 注册窗口事件
	mainWindow.OnWindowEvent(func(event *application.WindowEvent) {
		switch event.EventType {
		case application.WindowEventFocusLost:
			// 失去焦点时的处理
		case application.WindowEventFocusGained:
			// 获得焦点时的处理
		}
	})

	// 运行应用
	if err := wailsApp.Run(); err != nil {
		log.Fatalf("应用启动失败: %v", err)
		os.Exit(1)
	}
}
