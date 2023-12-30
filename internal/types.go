package internal

const (
	OCRobot = iota
	Tester
)

type Robot interface { // TODO: figure out architecture and what to send in channels
	Cmd() chan<- string
	Name() string
	Run()
}
