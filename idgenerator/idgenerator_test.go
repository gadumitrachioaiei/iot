package idgenerator

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestGet(t *testing.T) {
	m := make(map[string]bool)
	count := 100
	for _, id := range Get(count) {
		m[id] = true
	}
	if c := len(m); c != count {
		t.Fatalf("got %d unique ids, expected: %d", c, count)
	}
	id := `Asdasdasdsad"`
	fmt.Println(json.NewEncoder(os.Stdout).Encode(id))
	fmt.Println("second", `"`+id+`"`)
}
