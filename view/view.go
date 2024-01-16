package view

import (
	webview "github.com/webview/webview_go"
)

func Web(isDebug bool) {
	// f, _ := os.Open("frontend/dist/index.html")
	// b, _ := io.ReadAll(f)
	// f.Close()

	w := webview.New(true)
	defer w.Destroy()

	w.SetTitle("Talk with ChatGPT")
	w.SetSize(480, 920, webview.HintNone)
	// w.SetHtml(string(b))
	w.Navigate("http://localhost:5173/")
	w.Run()
}
