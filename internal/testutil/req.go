package testutil

type Request struct {
	Event     string
	GrayCount int
	UserId    string
	Attrs     map[string]string
}

func (r *Request) GetEvent() string {
	return r.Event
}

func (r *Request) GetGrayCount() int {
	return r.GrayCount
}
