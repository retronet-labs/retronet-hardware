// Package cpu implementa una mini-CPU didattica a 8 bit, assemblando i
// componenti di RetroNet: il register file e il Program Counter (sequenziali, a
// flip-flop), l'ALU di RetroNet Logic (combinatoria, a porte) e una memoria a
// byte. È il punto d'arrivo dello stack porte → … → CPU.
//
// L'ISA è descritto in docs/cpu-isa.md. Un'istruzione viene eseguita per ogni
// chiamata a [CPU.Step], che internamente pilota i fronti di clock dei
// componenti sequenziali. Gli unici calcoli aritmetici passano per l'ALU e per
// l'incrementatore del PC: nel datapath non si usano operatori aritmetici Go.
package cpu

import (
	"github.com/retronet-labs/retronet-logic/alu"
	"github.com/retronet-labs/retronet-logic/bit"
	"github.com/retronet-labs/retronet-logic/bus"
	"github.com/retronet-labs/retronet-logic/shifter"

	"github.com/retronet-labs/retronet-hardware/memory"
	"github.com/retronet-labs/retronet-hardware/pc"
	"github.com/retronet-labs/retronet-hardware/registerfile"
)

// Width è la larghezza di dati e indirizzi della mini-CPU.
const Width = 8

// Opcode (nibble alto del byte di opcode).
const (
	opNOP = 0x0
	opLDI = 0x1
	opMOV = 0x2
	opADD = 0x3
	opSUB = 0x4
	opAND = 0x5
	opOR  = 0x6
	opLD  = 0x7
	opST  = 0x8
	opJMP = 0x9
	opJZ  = 0xA
	opJC  = 0xB
	opSHL = 0xC
	opSHR = 0xD
	opHLT = 0xF
)

// CPU è lo stato della mini-CPU.
type CPU struct {
	regs *registerfile.File
	pc   *pc.PC
	mem  *memory.RAM

	// Registro di stato, modellato come campi: i flag prodotti dall'ALU.
	Carry bool
	Zero  bool

	// Halted diventa true dopo HLT.
	Halted bool
}

// New crea una mini-CPU con 4 registri da 8 bit collegata alla memoria mem.
func New(mem *memory.RAM) *CPU {
	return &CPU{
		regs: registerfile.New(4, Width),
		pc:   pc.New(Width),
		mem:  mem,
	}
}

// Reg restituisce il contenuto del registro i (0-3).
func (c *CPU) Reg(i int) byte { return byte(c.regs.Read(i).Uint()) }

// PC restituisce l'indirizzo corrente del program counter.
func (c *CPU) PC() byte { return byte(c.pc.Value().Uint()) }

// Run esegue istruzioni finché la CPU non si ferma (HLT) o finché non sono state
// eseguite maxSteps istruzioni (rete di sicurezza contro i loop infiniti).
func (c *CPU) Run(maxSteps int) {
	for i := 0; i < maxSteps && !c.Halted; i++ {
		c.Step()
	}
}

// Step esegue una singola istruzione (fetch → decode → execute).
func (c *CPU) Step() {
	if c.Halted {
		return
	}
	opcode := c.fetch()
	op := opcode >> 4
	rd := int((opcode >> 2) & 0x03)
	rs := int(opcode & 0x03)

	switch op {
	case opNOP:
		// niente
	case opLDI:
		c.writeReg(rd, byteBus(c.fetch()))
	case opMOV:
		c.writeReg(rd, c.regs.Read(rs))
	case opADD:
		c.aluOp(alu.Add, rd, rs, bit.Zero)
	case opSUB:
		c.aluOp(alu.Sub, rd, rs, bit.One) // cin=1: sottrazione semplice
	case opAND:
		c.aluOp(alu.And, rd, rs, bit.Zero)
	case opOR:
		c.aluOp(alu.Or, rd, rs, bit.Zero)
	case opLD:
		c.writeReg(rd, byteBus(c.mem.Read(c.fetch())))
	case opST:
		c.mem.Write(c.fetch(), byte(c.regs.Read(rd).Uint()))
	case opJMP:
		c.jump(c.fetch())
	case opJZ:
		addr := c.fetch()
		if c.Zero {
			c.jump(addr)
		}
	case opJC:
		addr := c.fetch()
		if c.Carry {
			c.jump(addr)
		}
	case opSHL:
		c.shiftOp(rd, true)
	case opSHR:
		c.shiftOp(rd, false)
	case opHLT:
		c.Halted = true
	}
}

// fetch legge il byte all'indirizzo del PC e poi incrementa il PC.
func (c *CPU) fetch() byte {
	b := c.mem.Read(c.PC())
	c.tickPC(zeroAddr, bit.Zero, bit.One) // PC = PC + 1
	return b
}

// aluOp esegue un'operazione ALU su Rd e Rs, scrive il risultato in Rd e
// aggiorna i flag.
func (c *CPU) aluOp(op alu.Op, rd, rs int, cin bit.Bit) {
	out, flags := alu.Compute(op, c.regs.Read(rd), c.regs.Read(rs), cin)
	c.writeReg(rd, out)
	c.Zero = flags.Zero.IsHigh()
	c.Carry = flags.Carry.IsHigh()
}

// shiftOp scorre Rd di una posizione (a sinistra o a destra) usando lo shifter a
// gate, riscrive Rd e aggiorna i flag: Carry = bit uscito, Zero = risultato nullo.
func (c *CPU) shiftOp(rd int, left bool) {
	out, carry := shifter.ShiftRight(c.regs.Read(rd))
	if left {
		out, carry = shifter.ShiftLeft(c.regs.Read(rd))
	}
	c.writeReg(rd, out)
	c.Carry = carry.IsHigh()
	c.Zero = out.Uint() == 0
}

// jump carica un nuovo indirizzo nel PC.
func (c *CPU) jump(addr byte) {
	c.tickPC(byteBus(addr), bit.One, bit.Zero)
}

// writeReg scrive data nel registro sel applicando un ciclo di clock completo.
func (c *CPU) writeReg(sel int, data bus.Bus) {
	c.regs.Step(sel, data, bit.One, bit.Zero)
	c.regs.Step(sel, data, bit.One, bit.One)
}

// tickPC applica un ciclo di clock completo al Program Counter.
func (c *CPU) tickPC(addr bus.Bus, load, inc bit.Bit) {
	c.pc.Step(addr, load, inc, bit.Zero)
	c.pc.Step(addr, load, inc, bit.One)
}

// byteBus converte un byte in un Bus a Width bit.
func byteBus(v byte) bus.Bus { return bus.FromUint(uint64(v), Width) }

// zeroAddr è l'indirizzo nullo usato quando addr è ininfluente (incremento).
var zeroAddr = bus.FromUint(0, Width)
