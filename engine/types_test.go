package engine

import "testing"

func TestAllTypesTableComplete(t *testing.T) {
	if len(typeData) != int(TypeCount) {
		t.Errorf("typeData covers %d types, but TypeCount is %d — update typeData in types.go when adding types", len(typeData), int(TypeCount))
	}
}

func TestTypeFromString_valid(t *testing.T) {
	for _, c := range typeData {
		t.Run(c.lower, func(t *testing.T) {
			got, err := TypeFromString(c.lower)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.typ {
				t.Errorf("got %v, want %v", got, c.typ)
			}
		})
	}
}

func TestTypeFromString_invalid(t *testing.T) {
	_, err := TypeFromString("banana")
	if err == nil {
		t.Error("expected error for unknown type, got nil")
	}
}

func TestTypeFromString_caseSensitive(t *testing.T) {
	_, err := TypeFromString("Fire")
	if err == nil {
		t.Error("expected error for uppercase input, got nil")
	}
}

func TestTypeString_known(t *testing.T) {
	for _, c := range typeData {
		t.Run(c.display, func(t *testing.T) {
			if got := c.typ.String(); got != c.display {
				t.Errorf("got %q, want %q", got, c.display)
			}
		})
	}
}

func TestTypeString_outOfRange(t *testing.T) {
	if got := Type(999).String(); got != "Unknown" {
		t.Errorf("got %q, want %q", got, "Unknown")
	}
}
