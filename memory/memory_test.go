package memory

import (
	"fmt"
	"testing"
)

func TestLetturaScrittura(t *testing.T) {
	m := New(256)
	if m.Size() != 256 {
		t.Fatalf("Size = %d, atteso 256", m.Size())
	}
	m.Write(0x10, 0xAB)
	if got := m.Read(0x10); got != 0xAB {
		t.Errorf("Read(0x10) = %#x, atteso 0xAB", got)
	}
	if got := m.Read(0x11); got != 0 {
		t.Errorf("cella non scritta = %#x, atteso 0", got)
	}
}

func TestLoad(t *testing.T) {
	m := New(256)
	m.Load([]byte{0x11, 0x22, 0x33})
	for addr, want := range []byte{0x11, 0x22, 0x33} {
		if got := m.Read(byte(addr)); got != want {
			t.Errorf("Read(%d) = %#x, atteso %#x", addr, got, want)
		}
	}
}

func ExampleRAM() {
	m := New(256)
	m.Write(5, 42)
	fmt.Println(m.Read(5))
	// Output: 42
}
