package view

import (
	"github.com/webview/webview"
)

func Web(isDebug bool) {
	// f, _ := os.Open("frontend/dist/index.html")
	// b, _ := io.ReadAll(f)
	// f.Close()

	w := webview.New(isDebug)
	defer w.Destroy()

	w.SetTitle("Talk with ChatGPT")
	w.SetSize(480, 490, webview.HintFixed)
	// w.SetHtml(string(b))
	w.Navigate("http://localhost:5173/")
	w.Run()
}
