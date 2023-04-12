package api

import (
	"fmt"
	"testing"
)

func TestVoice(t *testing.T) {
	c := New()

	c.BouyomiSpeaking("残りの人生を誰かと過ごしたいと思ったら、残りの人生をできるだけ早く始めたいと思うでしょう")
	fmt.Println("end")
	t.Log("#end")
}
