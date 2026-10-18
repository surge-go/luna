// Package ptr 提供值类型与指针之间的通用转换辅助。
package ptr

// To 返回指向 v 的指针。
func To[T any](v T) *T { return &v }

// From 解引用 p；p 为 nil 时返回 T 的零值。
func From[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// FromOr 解引用 p；p 为 nil 时返回 def。
func FromOr[T any](p *T, def T) T {
	if p == nil {
		return def
	}
	return *p
}
