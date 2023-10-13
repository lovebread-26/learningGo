package readfloat

import "fmt"

type TapeInterface interface {
	Play(string)
	Stop()
}

type TapePlayer struct {
	Batteries string
}

type TapeRecorder struct {
	Microphone int
}

func (t TapePlayer) Play(song string) {
	fmt.Println("Play the song:", song)
}

func (t TapePlayer) Stop() {
	fmt.Println("Stop the song")
}

func (t TapeRecorder) Play(song string) {
	fmt.Println("Start record the song:", song)
}

func (t TapeRecorder) Stop() {
	fmt.Println("Stop record the song")
}

func (t TapeRecorder) Record() {
	fmt.Println("Recording the song")
}

// 定义自己的error方法，返回string
func (t TapeRecorder) Error() string {
	// return "This is a common Error"
	return fmt.Sprintf("This is a common Error %#v", t)
}

// 定义自己的string方法，返回string
func (t TapeRecorder) String() string {
	return fmt.Sprintf("this is common string %#v:", t)
}

// error和string像上面这样定义，打印的一直是error
