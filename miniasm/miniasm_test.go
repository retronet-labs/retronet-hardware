package miniasm

import (
	"bytes"
	"testing"

	"github.com/retronet-labs/retronet-hardware/cpu"
	"github.com/retronet-labs/retronet-hardware/memory"
)

// La somma 1+2+3+4+5: confronto col codice macchina atteso e poi esecuzione
// sulla mini-CPU.
func TestSommaEndToEnd(t *testing.T) {
	src := `
        LDI R0, 0       ; somma
        LDI R1, 5       ; contatore
        LDI R2, 1       ; passo
loop:   ADD R0, R1
        SUB R1, R2
        JZ  end
        JMP loop
end:    ST  R0, 0x20
        HLT
`
	code, err := Assemble(src)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0x10, 0x00, 0x14, 0x05, 0x18, 0x01, 0x31, 0x46, 0xA0, 0x0C, 0x90, 0x06, 0x80, 0x20, 0xF0}
	if !bytes.Equal(code, want) {
		t.Fatalf("codice = % X\natteso = % X", code, want)
	}

	m := memory.New(256)
	m.Load(code)
	c := cpu.New(m)
	c.Run(1000)
	if got := m.Read(0x20); got != 15 {
		t.Errorf("mem[0x20] = %d, atteso 15", got)
	}
}

// Una subroutine chiamata con CALL che raddoppia un valore tramite SHL.
func TestCallRetEndToEnd(t *testing.T) {
	src := `
        LDI R0, 21
        CALL doppio
        ST  R0, 0x30
        HLT
doppio: SHL R0          ; R0 = R0 * 2
        RET
`
	code, err := Assemble(src)
	if err != nil {
		t.Fatal(err)
	}
	m := memory.New(256)
	m.Load(code)
	c := cpu.New(m)
	c.Run(1000)
	if got := m.Read(0x30); got != 42 {
		t.Errorf("mem[0x30] = %d, atteso 42 (21 raddoppiato dalla subroutine)", got)
	}
}

func TestErrori(t *testing.T) {
	casi := []string{
		"FOO R0, R1",  // istruzione sconosciuta
		"ADD R0",      // operandi insufficienti
		"ADD R0, R9",  // registro non valido
		"LDI R0, 999", // immediato fuori da 8 bit
		"JMP manca",   // etichetta non definita
	}
	for _, src := range casi {
		if _, err := Assemble(src); err == nil {
			t.Errorf("atteso errore per %q", src)
		}
	}
}
