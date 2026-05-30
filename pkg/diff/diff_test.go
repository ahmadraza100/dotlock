package diff

import (
	"reflect"
	"testing"
)

func TestDiffMaps(t *testing.T) {
	a := map[string][]byte{"A": []byte("1"), "B": []byte("2")}
	b := map[string][]byte{"B": []byte("2"), "C": []byte("3")}
	out := Maps(a, b)
	expected := []string{"- A", "  B", "+ C"}
	if !reflect.DeepEqual(out, expected) {
		t.Fatalf("unexpected diff: %v", out)
	}
}
