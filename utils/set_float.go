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

func (flt *float64Set) copy() *float64Set {
	new := float64Set{
		values: map[float64]exist{},
	}

	new.Append(flt.Values()...)

	return &new
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

func (flt *float64Set) Len() int {
	return len(flt.values)
}

func checkFloat64Set(b Set) *float64Set {
	if b.Len() < 1 {
		return nil
	}

	other, ok := b.(*float64Set)

	if !ok {
		panic("invalid type for StringSet")
	}

	return other
}

func (flt *float64Set) Add(b Set) Set {
	new := flt.copy()

	if other := checkFloat64Set(b); other != nil {
		for k := range other.values {
			new.values[k] = exist{}
		}
	}

	return new
}

func (flt *float64Set) Join(b Set) Set {
	new := flt.copy()

	if other := checkFloat64Set(b); other != nil {
		for k := range other.values {
			if !flt.Exist(k) {
				delete(new.values, k)
			}
		}
	}

	return new
}

func (flt *float64Set) Sub(b Set) Set {
	new := flt.copy()

	if other := checkFloat64Set(b); other != nil {
		for k := range other.values {
			if flt.Exist(k) {
				delete(new.values, k)
			}
		}
	}

	return new
}

func (flt *float64Set) Contain(b Set) bool {
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
