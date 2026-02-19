package core

type Response struct {
	Code int
	Msg  map[string]string
}

func (r *Response) SetCode(code int) bool {
	r.Code = code
	return true
}

func (r *Response) SetMsg(key, value string) bool {
	if r.Msg == nil {
		r.Msg = make(map[string]string)
	}
	r.Msg[key] = value

	return true
}

func (r *Response) GetMsg(key string) string {
	if r.Msg == nil {
		return ""
	}
	return r.Msg[key]
}
