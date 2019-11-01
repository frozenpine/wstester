package utils

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
	// Len set length
	Len() int
}
