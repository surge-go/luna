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

// BoolValue 返回布尔指针的值，如果为 nil 则返回默认值。
func BoolValue(p *bool, def bool) bool {
	return FromOr(p, def)
}

// Int16Value 返回 int16 指针的值，如果为 nil 则返回默认值。
func Int16Value(p *int16, def int16) int16 {
	return FromOr(p, def)
}

// IntValue 返回 int 指针的值，如果为 nil 则返回默认值。
func IntValue(p *int, def int) int {
	return FromOr(p, def)
}

// StringValue 返回字符串指针的值，如果为 nil 则返回默认值。
func StringValue(p *string, def string) string {
	return FromOr(p, def)
}
