package cpu

import (
	"testing"

	"github.com/retronet-labs/retronet-hardware/memory"
)

func newCPU(program []byte) (*CPU, *memory.RAM) {
	m := memory.New(256)
	m.Load(program)
	return New(m), m
}

func TestLDIeMOV(t *testing.T) {
	c, _ := newCPU([]byte{
		0x10, 0x2A, // LDI R0, 42
		0x24, // MOV R1, R0
		0xF0, // HLT
	})
	c.Run(10)
	if c.Reg(0) != 42 || c.Reg(1) != 42 {
		t.Errorf("R0=%d R1=%d, attesi 42 42", c.Reg(0), c.Reg(1))
	}
}

func TestADDFlags(t *testing.T) {
	c, _ := newCPU([]byte{
		0x10, 0xC8, // LDI R0, 200
		0x14, 0x64, // LDI R1, 100
		0x31, // ADD R0, R1  -> 300 & 0xFF = 44, carry
		0xF0,
	})
	c.Run(10)
	if c.Reg(0) != 44 {
		t.Errorf("R0=%d, atteso 44", c.Reg(0))
	}
	if !c.Carry || c.Zero {
		t.Errorf("flag: carry=%v zero=%v, attesi carry=true zero=false", c.Carry, c.Zero)
	}
}

func TestSUBZero(t *testing.T) {
	c, _ := newCPU([]byte{
		0x10, 0x05, // LDI R0, 5
		0x14, 0x05, // LDI R1, 5
		0x41, // SUB R0, R1  -> 0, zero, nessun prestito (carry=1)
		0xF0,
	})
	c.Run(10)
	if c.Reg(0) != 0 || !c.Zero || !c.Carry {
		t.Errorf("R0=%d zero=%v carry=%v, attesi 0 true true", c.Reg(0), c.Zero, c.Carry)
	}
}

func TestLoadStore(t *testing.T) {
	c, m := newCPU([]byte{
		0x10, 0x37, // LDI R0, 0x37
		0x80, 0x20, // ST R0, 0x20
		0x74, 0x20, // LD R1, 0x20
		0xF0,
	})
	c.Run(10)
	if got := m.Read(0x20); got != 0x37 {
		t.Errorf("mem[0x20]=%#x, atteso 0x37", got)
	}
	if c.Reg(1) != 0x37 {
		t.Errorf("R1=%#x, atteso 0x37", c.Reg(1))
	}
}

func TestJMPSaltaIstruzione(t *testing.T) {
	c, _ := newCPU([]byte{
		0x10, 0x01, // LDI R0, 1
		0x90, 0x06, // JMP 0x06
		0x10, 0x63, // LDI R0, 99  (deve essere saltata)
		0xF0, // 0x06: HLT
	})
	c.Run(10)
	if c.Reg(0) != 1 {
		t.Errorf("R0=%d, atteso 1 (JMP deve saltare la LDI 99)", c.Reg(0))
	}
}

func TestCallRet(t *testing.T) {
	c, m := newCPU([]byte{
		0x10, 0x00, // 0x00 LDI R0, 0
		0xE1, 0x08, // 0x02 CALL 0x08
		0x80, 0x20, // 0x04 ST R0, 0x20
		0xF0,       // 0x06 HLT
		0x00,       // 0x07 (padding)
		0x10, 0x2A, // 0x08 LDI R0, 42   (subroutine)
		0xE0, // 0x0A RET
	})
	c.Run(100)
	if got := m.Read(0x20); got != 42 {
		t.Errorf("mem[0x20]=%d, atteso 42 (CALL deve eseguire la subroutine e RET tornare)", got)
	}
	if c.SP() != 0xFF {
		t.Errorf("SP=%#x, atteso 0xFF (pila ripristinata dopo RET)", c.SP())
	}
}

func TestSHLeSHR(t *testing.T) {
	// SHL: 0x81 << 1 = 0x02, carry = vecchio MSB = 1.
	c, _ := newCPU([]byte{
		0x10, 0x81, // LDI R0, 0x81
		0xC0, // SHL R0
		0xF0,
	})
	c.Run(10)
	if c.Reg(0) != 0x02 || !c.Carry {
		t.Errorf("SHL: R0=%#x carry=%v, attesi 0x02 true", c.Reg(0), c.Carry)
	}

	// SHR: 0x01 >> 1 = 0x00, carry = vecchio LSB = 1, zero = 1.
	c, _ = newCPU([]byte{
		0x10, 0x01, // LDI R0, 1
		0xD0, // SHR R0
		0xF0,
	})
	c.Run(10)
	if c.Reg(0) != 0x00 || !c.Carry || !c.Zero {
		t.Errorf("SHR: R0=%#x carry=%v zero=%v, attesi 0 true true", c.Reg(0), c.Carry, c.Zero)
	}
}

// Programma completo (vedi docs/cpu-isa.md): somma 1+2+3+4+5 = 15 in mem[0x20].
func TestProgrammaSomma(t *testing.T) {
	prog := []byte{
		0x10, 0x00, // LDI R0, 0   (somma)
		0x14, 0x05, // LDI R1, 5   (contatore)
		0x18, 0x01, // LDI R2, 1   (passo)
		0x31,       // loop: ADD R0, R1
		0x46,       // SUB R1, R2
		0xA0, 0x0C, // JZ end (0x0C)
		0x90, 0x06, // JMP loop (0x06)
		0x80, 0x20, // end: ST R0, 0x20
		0xF0, // HLT
	}
	c, m := newCPU(prog)
	c.Run(100)
	if !c.Halted {
		t.Fatal("la CPU non si è fermata")
	}
	if got := m.Read(0x20); got != 15 {
		t.Errorf("mem[0x20]=%d, atteso 15", got)
	}
}
