// Command cpu esegue un programma sulla mini-CPU di RetroNet Hardware e ne mostra
// il risultato. Il programma somma 1+2+3+4+5 e salva 15 in mem[0x20].
//
// Esecuzione:
//
//	go run ./examples/cpu
package main

import (
	"fmt"

	"github.com/retronet-labs/retronet-hardware/cpu"
	"github.com/retronet-labs/retronet-hardware/memory"
)

func main() {
	// Codice macchina (vedi docs/cpu-isa.md):
	//   LDI R0,0 ; LDI R1,5 ; LDI R2,1
	//   loop: ADD R0,R1 ; SUB R1,R2 ; JZ end ; JMP loop
	//   end:  ST R0,0x20 ; HLT
	prog := []byte{
		0x10, 0x00,
		0x14, 0x05,
		0x18, 0x01,
		0x31,
		0x46,
		0xA0, 0x0C,
		0x90, 0x06,
		0x80, 0x20,
		0xF0,
	}

	m := memory.New(256)
	m.Load(prog)
	c := cpu.New(m)

	fmt.Println("Mini-CPU RetroNet — somma 1+2+3+4+5")
	c.Run(100)
	fmt.Printf("fermata    : %v\n", c.Halted)
	fmt.Printf("R0         : %d\n", c.Reg(0))
	fmt.Printf("mem[0x20]  : %d\n", m.Read(0x20))
}
