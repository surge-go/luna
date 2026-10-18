package ptr

import "testing"

func TestTo(t *testing.T) {
	p := To(42)
	if p == nil {
		t.Fatal("To(42) = nil, want non-nil")
	}
	if *p != 42 {
		t.Fatalf("To(42) = %d, want 42", *p)
	}
}

func TestTo_String(t *testing.T) {
	p := To("admin")
	if p == nil || *p != "admin" {
		t.Fatalf("To(\"admin\") = %v, want *admin", p)
	}
}

func TestFrom(t *testing.T) {
	v := 42
	if got := From(&v); got != 42 {
		t.Fatalf("From(&42) = %d, want 42", got)
	}
}

func TestFrom_Nil(t *testing.T) {
	var p *int
	if got := From(p); got != 0 {
		t.Fatalf("From(nil) = %d, want 0", got)
	}
}

func TestFrom_StringNil(t *testing.T) {
	var p *string
	if got := From(p); got != "" {
		t.Fatalf("From(nil string) = %q, want empty", got)
	}
}

func TestFromOr(t *testing.T) {
	v := 42
	if got := FromOr(&v, 99); got != 42 {
		t.Fatalf("FromOr(&42, 99) = %d, want 42", got)
	}
}

func TestFromOr_Nil(t *testing.T) {
	var p *int
	if got := FromOr(p, 99); got != 99 {
		t.Fatalf("FromOr(nil, 99) = %d, want 99", got)
	}
}
