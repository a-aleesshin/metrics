package resettest

import "testing"

func TestContainerReset(t *testing.T) {
	str := "hello"
	num := 42

	c := &Container{
		Num:    7,
		Str:    "value",
		Flag:   true,
		Ratio:  1.5,
		StrPtr: &str,
		NumPtr: &num,
		Items:  []string{"a", "b", "c"},
		Index:  map[string]int{"a": 1, "b": 2},
		Child:  Child{Value: 10},
		ChildP: &Child{Value: 20},
	}

	itemsCap := cap(c.Items)

	c.Reset()

	if c.Num != 0 || c.Str != "" || c.Flag || c.Ratio != 0 {
		t.Fatalf("primitives not reset: %+v", c)
	}

	if str != "" {
		t.Fatalf("expected *StrPtr reset to empty string, got %q", str)
	}

	if num != 0 {
		t.Fatalf("expected *NumPtr reset to 0, got %d", num)
	}

	if len(c.Items) != 0 {
		t.Fatalf("expected empty slice, got %v", c.Items)
	}

	if cap(c.Items) != itemsCap {
		t.Fatalf("expected slice capacity %d preserved, got %d", itemsCap, cap(c.Items))
	}

	if len(c.Index) != 0 {
		t.Fatalf("expected cleared map, got %v", c.Index)
	}

	if c.Child.Value != 0 {
		t.Fatalf("expected nested struct reset, got %d", c.Child.Value)
	}

	if c.ChildP == nil || c.ChildP.Value != 0 {
		t.Fatalf("expected nested pointer struct reset, got %+v", c.ChildP)
	}
}

func TestContainerReset_NilReceiverAndNilFields(t *testing.T) {
	var nilContainer *Container
	nilContainer.Reset()

	empty := &Container{}
	empty.Reset()

	if empty.Items != nil {
		t.Fatalf("expected nil slice to stay nil, got %v", empty.Items)
	}
}
