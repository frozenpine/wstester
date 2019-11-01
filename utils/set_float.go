package utils

// Float64Set set for float64
type Float64Set interface {
	Set
	Append(...float64)
	Values() []float64
	Exist(float64) bool
}

type float64Set struct {
	values map[float64]exist
}

func (flt *float64Set) Append(elements ...float64) {
	for _, num := range elements {
		flt.values[num] = exist{}
	}
}

func (flt *float64Set) Values() []float64 {
	var keys []float64

	for k := range flt.values {
		keys = append(keys, k)
	}

	return keys
}

func (flt *float64Set) Exist(v float64) bool {
	_, exist := flt.values[v]

	return exist
}

func checkFloat64Set(b Set) *float64Set {
	if _, isNil := b.(NilSet); isNil {
		return nil
	}

	other, ok := b.(float64Set)

	if !ok {
		panic("invalid type for StringSet")
	}

	return &other
}

func (flt float64Set) Add(b Set) Set {
	if other := checkFloat64Set(b); other != nil {
		for k := range other.values {
			flt.values[k] = exist{}
		}
	}

	return &flt
}

func (flt float64Set) Join(b Set) Set {
	if other := checkFloat64Set(b); other != nil {
		for k := range other.values {
			if !flt.Exist(k) {
				delete(flt.values, k)
			}
		}
	}

	return &flt
}

func (flt float64Set) Sub(b Set) Set {
	if other := checkFloat64Set(b); other != nil {
		for k := range other.values {
			if flt.Exist(k) {
				delete(flt.values, k)
			}
		}
	}

	return &flt
}

func (flt float64Set) Contain(b Set) bool {
	contain := true

	if other := checkFloat64Set(b); other != nil {
		for k := range other.values {
			if !flt.Exist(k) {
				contain = false
				break
			}
		}
	}

	return contain
}

// NewFloat64Set convert a float64 slice to Float64Set
func NewFloat64Set(ele []float64) Float64Set {
	set := float64Set{
		values: make(map[float64]exist),
	}

	if ele != nil {
		set.Append(ele...)
	}

	return &set
}
