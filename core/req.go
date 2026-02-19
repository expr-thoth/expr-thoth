package core

type Request interface {
	// 事件类型, 例如登录事件, 支付事件等, 由调用方根据实际场景定义
	GetEvent() string

	// 灰度值, 取值范围为[0, 10000), 例如可以使用UserId的Hash值对10000取模的结果作为灰度值, 从而实现基于用户维度的灰度控制
	GetGrayCount() int
}

// 策略初始化的Request对象
type InitRequest struct{}

func (r InitRequest) GetEvent() string {
	return "event_init_policy"
}

func (r InitRequest) GetGrayCount() int {
	return 0
}
