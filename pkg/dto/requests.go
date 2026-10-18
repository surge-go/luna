package dto

import (
	stderrors "errors"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	// DateTimeLayout 是请求时间字段使用的年月日时分秒格式。
	DateTimeLayout = "2006-01-02 15:04:05"
	// DefaultTimezone 是时间范围请求未指定时区时使用的北京时区。
	DefaultTimezone = "Asia/Shanghai"
	// MaxPageSize 是单次分页查询允许请求的最大记录数。
	MaxPageSize = 100
	defaultPage = 1
)

var (
	// ErrInvalidTimezone 表示请求的时区名称无法识别。
	ErrInvalidTimezone = stderrors.New("invalid timezone")
)

// IDReq 携带单个 ID 的查询或操作请求。
type IDReq struct {
	ID int64 `json:"id" form:"id" uri:"id" binding:"required,gt=0"`
}

// IdsReq 携带 ID 集合的批量操作请求。
type IdsReq struct {
	Ids []int64 `json:"ids" form:"ids" binding:"required,min=1,dive,gt=0"`
}

// StatusReq 携带启用状态过滤的查询请求。
// Status 使用指针类型以区分「未传」与「显式传 0（禁用）」两种语义。
type StatusReq struct {
	Status *int8 `json:"status" form:"status" binding:"omitempty,oneof=0 1"`
}

// SortReq 携带单字段排序参数。
// Order 使用指针类型以区分「未传（默认降序）」与「显式传 0」两种语义。
type SortReq struct {
	Field string `json:"field" form:"field"`
	// Order 排序方向：0 降序（desc），1 升序（asc）。
	Order *int `json:"order" form:"order" binding:"omitempty,oneof=0 1"`
}

// Desc 返回是否按降序排序（Order 未传或为 0 时表示 desc，为 1 时表示 asc）。
func (r *SortReq) Desc() bool {
	return r.Order == nil || *r.Order == 0
}

// PageReq 携带分页参数。
// Page/Size 小于 1 或超过允许范围时，由 Offset/Limit 规范化为有效值。
type PageReq struct {
	Page int `json:"page" form:"page"`
	Size int `json:"size" form:"size"`
}

// Offset 返回分页偏移量。Page 小于 1 时按第 1 页处理。
func (r *PageReq) Offset() int {
	limit := r.Limit()
	page := r.Page
	if page < 1 {
		page = defaultPage
	}
	if page-1 > math.MaxInt/limit {
		return math.MaxInt
	}
	return (page - 1) * limit
}

// Limit 返回每页条数。Size 小于 1 时返回 1，超过 MaxPageSize 时返回 MaxPageSize。
func (r *PageReq) Limit() int {
	if r.Size < 1 {
		return 1
	}
	if r.Size > MaxPageSize {
		return MaxPageSize
	}
	return r.Size
}

// TimeRangeReq 携带时间范围查询参数。
// From/To 格式为 YYYY-MM-DD HH:mm:ss；Timezone 使用 IANA 时区名称，未传时默认为 Asia/Shanghai。
type TimeRangeReq struct {
	From     string `json:"from" form:"from"`
	To       string `json:"to" form:"to"`
	Timezone string `json:"timezone" form:"timezone"`
}

// Parse 解析时间范围，返回规范化后的起止时间。
// From/To 任一为空字符串表示对应边界不限制（返回 nil）。
func (r *TimeRangeReq) Parse() (*time.Time, *time.Time, error) {
	location, err := r.location()
	if err != nil {
		return nil, nil, err
	}

	from, err := parseOptionalTime("from", r.From, location)
	if err != nil {
		return nil, nil, err
	}

	to, err := parseOptionalTime("to", r.To, location)
	if err != nil {
		return nil, nil, err
	}

	if from != nil && to != nil && from.After(*to) {
		return nil, nil, fmt.Errorf("from time must not be after to time")
	}
	return from, to, nil
}

func (r *TimeRangeReq) location() (*time.Location, error) {
	timezone := strings.TrimSpace(r.Timezone)
	if timezone == "" {
		timezone = DefaultTimezone
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("%w %q: %v", ErrInvalidTimezone, timezone, err)
	}
	return location, nil
}

func parseOptionalTime(field, value string, location *time.Location) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	parsed, err := time.ParseInLocation(DateTimeLayout, value, location)
	if err != nil {
		return nil, fmt.Errorf("parse %s time: %w", field, err)
	}
	return &parsed, nil
}
