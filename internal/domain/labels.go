package domain

type Labels map[string]string

func (l Labels) Copy() Labels {
	out := Labels{}
	for k, v := range l {
		out[k] = v
	}
	return out
}
func (l Labels) Get(key string) (string, bool) { v, ok := l[key]; return v, ok }
func (l Labels) Set(key, value string) Labels  { out := l.Copy(); out[key] = value; return out }
func (l Labels) Without(key string) Labels     { out := l.Copy(); delete(out, key); return out }
func (l Labels) Empty() bool                   { return len(l) == 0 }
func (l Labels) Len() int                      { return len(l) }
func (l Labels) Keys() []string {
	out := make([]string, 0, len(l))
	for k := range l {
		out = append(out, k)
	}
	return out
}
func (l Labels) Has(key string) bool { _, ok := l[key]; return ok }
func (l Labels) Merge(other Labels) Labels {
	out := l.Copy()
	for k, v := range other {
		out[k] = v
	}
	return out
}
func (l Labels) Clone() Labels { return l.Copy() }
func (l Labels) Equal(other Labels) bool {
	if len(l) != len(other) {
		return false
	}
	for k, v := range l {
		if other[k] != v {
			return false
		}
	}
	return true
}
