package utils

// NilSet is an empty set
type NilSet Set

type exist struct{}

// Set uniqueue
type Set interface {
	// Add return an union set
	Add(Set) Set
	// Join 计算交集
	Join(Set) Set
	// Sub 计算差集
	Sub(Set) Set
	// Contain 是否为子集
	Contain(Set) bool
}

// StringSet set for string
type StringSet interface {
	// Append append strings to Set
	Append(...string)
	// Values get string slice
	Values() []string
	// Exist test string exist in set
	Exist(string) bool
}

type stringSet struct {
	values map[string]exist
}

func (ss *stringSet) Append(elements ...string) {
	for _, str := range elements {
		ss.values[str] = exist{}
	}
}

func (ss *stringSet) Values() []string {
	var keys []string

	for k := range ss.values {
		keys = append(keys, k)
	}

	return keys
}

func (ss *stringSet) Exist(v string) bool {
	_, exist := ss.values[v]

	return exist
}

func checkStringSet(b Set) *stringSet {
	if _, isNil := b.(NilSet); isNil {
		return nil
	}

	other, ok := b.(stringSet)

	if !ok {
		panic("invalid type for StringSet")
	}

	return &other
}

func (ss stringSet) Add(b Set) Set {
	if other := checkStringSet(b); other != nil {
		for k := range other.values {
			ss.values[k] = exist{}
		}
	}

	return &ss
}

func (ss stringSet) Join(b Set) Set {
	if other := checkStringSet(b); other != nil {
		for k := range other.values {
			if !ss.Exist(k) {
				delete(ss.values, k)
			}
		}
	}

	return &ss
}

func (ss stringSet) Sub(b Set) Set {
	if other := checkStringSet(b); other != nil {
		for k := range other.values {
			if ss.Exist(k) {
				delete(ss.values, k)
			}
		}
	}

	return &ss
}

func (ss stringSet) Contain(b Set) bool {
	var contain bool = true

	if other := checkStringSet(b); other != nil {
		for k := range other.values {
			if !ss.Exist(k) {
				contain = false
				break
			}
		}
	}

	return contain
}

// NewStringSet convert a string slice to StringSet
func NewStringSet(ele []string) StringSet {
	set := stringSet{
		values: make(map[string]exist),
	}

	set.Append(ele...)

	return &set
}
