package utils

// StringSet set for string
type StringSet interface {
	Set
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

func (ss *stringSet) copy() *stringSet {
	new := stringSet{
		values: map[string]exist{},
	}

	new.Append(ss.Values()...)

	return &new
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

func (ss *stringSet) Len() int {
	return len(ss.values)
}

func checkStringSet(b Set) *stringSet {
	if b.Len() < 1 {
		return nil
	}

	other, ok := b.(*stringSet)

	if !ok {
		panic("invalid type for StringSet")
	}

	return other
}

func (ss *stringSet) Add(b Set) Set {
	new := ss.copy()

	if other := checkStringSet(b); other != nil {
		for k := range other.values {
			new.values[k] = exist{}
		}
	}

	return new
}

func (ss *stringSet) Join(b Set) Set {
	new := ss.copy()

	if other := checkStringSet(b); other != nil {
		for k := range other.values {
			if !ss.Exist(k) {
				delete(new.values, k)
			}
		}
	}

	return new
}

func (ss *stringSet) Sub(b Set) Set {
	new := ss.copy()

	if other := checkStringSet(b); other != nil {
		for k := range other.values {
			if ss.Exist(k) {
				delete(new.values, k)
			}
		}
	}

	return new
}

func (ss *stringSet) Contain(b Set) bool {
	contain := true

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

	if ele != nil {
		set.Append(ele...)
	}

	return &set
}
