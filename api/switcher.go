package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

var (
	//go:embed utils-stopwords.txt
	stopwords string

	LoadText  string
	STOPWORDS []string
)

func init() {
	STOPWORDS = strings.Split(stopwords, "\n")
}

func (p *Client) Switcher(c echo.Context, message string) (isSwitch bool, err error) {
	isSwitch = true
	q := strings.ToLower(message)

	// Has not query
	if message == "" {
		return isSwitch, c.JSON(http.StatusBadRequest, map[string]any{
			"code":  "error bad request, has not request message",
			"error": errors.New("has not query"),
		})
	}

	// Reset...
	if strings.HasPrefix(q, "reset") ||
		strings.HasPrefix(q, "リセット") ||
		strings.HasPrefix(q, "りせっと") ||
		strings.HasPrefix(q, "履歴リセット") ||
		strings.HasPrefix(q, "履歴りせっと") {
		// System sayを残す
		ResetHistories()
		return isSwitch, c.JSON(http.StatusOK, map[string]any{
			"code":   "success",
			"answer": "success, reset histories",
		})
	}

	// Change model
	if q == "モデル" ||
		q == "モデル。" {
		go p.BouyomiSpeaking(fmt.Sprintf("現在のモデルは%sです。", p.UseModel()))
		return isSwitch, c.JSON(http.StatusOK, map[string]any{
			"code":   "success",
			"answer": "success, use model: " + p.UseModel(),
		})
	} else if strings.HasPrefix(q, "モデル") {
		temp := strings.Replace(q, "モデル", "", 1)
		if strings.HasPrefix(temp, "3.5") {
			p.SetModel(MODELGPT3_5)
			go p.BouyomiSpeaking("モデル3.5に切り替え。")
			return isSwitch, c.JSON(http.StatusOK, map[string]any{
				"code":   "success",
				"answer": "success, set model: " + p.UseModel(),
			})
		} else if strings.HasPrefix(temp, "4.0") {
			p.SetModel(MODELGPT4)
			go p.BouyomiSpeaking("モデル4.0に切り替え。")
			return isSwitch, c.JSON(http.StatusOK, map[string]any{
				"code":   "success",
				"answer": "success, set model: " + p.UseModel(),
			})
		}
	}

	// Memo...
	if strings.HasPrefix(q, "メモ") {
		temp := strings.Replace(q, "メモを", "", 1)
		temp = strings.Replace(temp, "メモ", "", 1)
		temp = strings.Replace(temp, "、", "", 1)
		// 履歴の読み込み
		if strings.HasPrefix(temp, "読み込み") {
			temp = strings.Replace(temp, "読み込み", "", 1)
			temp = strings.Replace(temp, "。", "", -1)
			temp = strings.Replace(temp, "、", "", -1)
			text, err := _search(temp)
			if err != nil {
				go p.BouyomiSpeaking(fmt.Sprintf("「%s」が含むファイルがみつからなかったよ。", temp))
				return isSwitch, c.JSON(http.StatusOK, map[string]any{
					"code":   "error",
					"answer": err.Error(),
				})
			}

			tmp := fmt.Sprintf("「%s」からは特定出来なかったから以下の特定キーワードを抽出したよ。", temp)
			if strings.HasPrefix(text, tmp) {
				go p.BouyomiSpeaking(tmp)
			} else {
				go p.BouyomiSpeaking(fmt.Sprintf("「%s」を含むファイルを見つけたよ。", temp))
			}

			LoadText = text
			return isSwitch, c.JSON(http.StatusOK, map[string]any{
				"code":   "success",
				"answer": text,
			})
		}

		// 履歴の書き込み
		say, err := Save(his)
		if err != nil {
			return isSwitch, c.JSON(http.StatusOK, map[string]any{
				"code":   "error",
				"answer": err.Error(),
			})
		}

		say = fmt.Sprintf("%s、 履歴をリセットしました。", say)

		// System sayを残す
		ResetHistories()

		go p.BouyomiSpeaking(say)

		return isSwitch, c.JSON(http.StatusOK, map[string]any{
			"code":   "success",
			"answer": say,
		})
	}

	// 現在の配列最後を読み上げる
	if strings.HasPrefix(q, "履歴復習") ||
		strings.HasPrefix(q, "履歴復讐") ||
		strings.HasPrefix(q, "履歴復習。") ||
		strings.HasPrefix(q, "履歴復讐。") {
		var (
			say        []string
			returntext []string
		)
		for i := 0; i < len(his); i++ {
			returntext = append(returntext, his[i].Content)
			if his[i].Role != "assistant" {
				continue
			}
			say = append(say, his[i].Content)
		}

		go p.BouyomiSpeaking(strings.Join(say, "\n\n"))

		return isSwitch, c.JSON(http.StatusOK, map[string]any{
			"code":   "success",
			"answer": strings.Join(returntext, "\n\n"),
		})
	}

	// 現在の配列最後を読み上げる
	if strings.HasPrefix(q, "読み上げ") {
		if LoadText != "" {
			go p.BouyomiSpeaking(LoadText)

			return isSwitch, c.JSON(http.StatusOK, map[string]any{
				"code": "success",
			})
		}

		go p.BouyomiSpeaking(his[len(his)-1].Content)

		return isSwitch, c.JSON(http.StatusOK, map[string]any{
			"code": "success",
		})
	}

	return false, nil
}

var Dir = "./_data"

// _search ファイル走査 該当のサーチクエリが含まれる先頭のファイルを返す単純な関数
func _search(q string) (string, error) {
	var (
		match []string
		str   []string
	)

	files, _ := os.ReadDir(Dir)
	for _, file := range files {
		f, _ := os.Open(filepath.Join(Dir, file.Name()))
		buf, err := io.ReadAll(f)
		if err != nil {
			log.Err(err).Msg("")
			continue
		}
		f.Close()

		if strings.Contains(string(buf), q) {
			match = append(match, string(buf))
		}

		str = append(str, string(buf))
	}

	if len(match) == 1 {
		return match[0], nil
	}

	if len(match) <= 0 {
		// 照合ファイルが何もない
		if len(str) <= 0 {
			return "has not file", fmt.Errorf("has not file")
		}
	} else {
		// 探したいキーワードに合致するファイルが多数ある
		str = match
	}

	// 照合合致ファイルがない場合は、ファイル群のテキストからキーワードを返す
	keywords := _toKeywords(str)
	return fmt.Sprintf("「%s」からは特定出来なかったから以下の特定キーワードを抽出したよ。\n\n%s", q, strings.Join(keywords, "\n")), nil
}

func _toKeywords(str []string) []string {
	var temp string
	temp = strings.ReplaceAll(strings.Join(str, " "), "、", " ")
	temp = strings.ReplaceAll(temp, "。", " ")
	temp = strings.ReplaceAll(temp, "！", " ")
	temp = strings.ReplaceAll(temp, "？", " ")
	temp = strings.ReplaceAll(temp, "「", " ")
	temp = strings.ReplaceAll(temp, "」", " ")
	words := strings.Split(temp, " ")

	var keywords []string
	for i := 0; i < len(words); i++ {
		if !_stopwords(words[i]) {
			keywords = append(keywords, words[i])
		}
	}

	return _removeDuplicates(keywords)
}

func _stopwords(q string) bool {
	for _, word := range STOPWORDS {
		if q == word {
			return true
		}
	}
	return false
}

func _removeDuplicates(words []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range words {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
