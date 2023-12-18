package api

import (
	"fmt"
	"testing"

	gogpt "github.com/sashabaranov/go-openai"
)

func TestMeCab(t *testing.T) {

}

func TestVoice(t *testing.T) {
	c := New(gogpt.GPT3Dot5Turbo)

	c.BouyomiSpeaking("残りの人生を誰かと過ごしたいと思ったら、残りの人生をできるだけ早く始めたいと思うでしょう")
	fmt.Println("end")
	t.Log("#end")
}
