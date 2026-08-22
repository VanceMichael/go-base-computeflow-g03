package domain

type BatchItemResult struct {
	Key      string
	Accepted bool
	Code     string
	Message  string
}
type BatchResult struct {
	Items    []BatchItemResult
	Accepted int
	Rejected int
	Atomic   bool
}

func (r *BatchResult) Add(key string, accepted bool, code, message string) {
	r.Items = append(r.Items, BatchItemResult{Key: key, Accepted: accepted, Code: code, Message: message})
	if accepted {
		r.Accepted++
	} else {
		r.Rejected++
	}
}
func (r BatchResult) Complete() bool { return r.Rejected == 0 }
func (r BatchResult) FailedKeys() []string {
	out := make([]string, 0, r.Rejected)
	for _, item := range r.Items {
		if !item.Accepted {
			out = append(out, item.Key)
		}
	}
	return out
}
