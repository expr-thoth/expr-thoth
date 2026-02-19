package ds

type Slice struct{}

func (s Slice) CreateSlice() []any {
	return make([]any, 0)
}