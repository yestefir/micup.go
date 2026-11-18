package voice_test

import (
	"fmt"

	"github.com/yestefir/micup.go/voice"
)

func ExampleScaleVolume() {
	frame := []int16{1000, -1000, 30000}
	voice.ScaleVolume(frame, 0.5)
	fmt.Println(frame)
	voice.ScaleVolume(frame, 4)
	fmt.Println(frame)
	// Output:
	// [500 -500 15000]
	// [2000 -2000 32767]
}

func ExamplePlayer_SetVolume() {
	player := voice.NewPlayer(nil)
	fmt.Println(player.SetVolume(0.8))
	fmt.Println(player.SetVolume(10))
	fmt.Println(player.Volume())
	// Output:
	// 0.8
	// 2
	// 2
}
