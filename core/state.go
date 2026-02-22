package core

// 策略引擎内部状态
type ThothState struct {
	IsReturn    bool
	IsExitGroup bool
	IsHalt      bool
}

var DefaultThothState = &ThothState{}
