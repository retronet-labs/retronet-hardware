// Package i6502 adatta la ALU a porte di RetroNet Logic alla semantica
// aritmetico-logica del MOS Technology 6502 NMOS.
//
// Il 6502 usa un accumulatore a 8 bit, flag Carry come "nessun prestito" nelle
// sottrazioni e una modalita' decimale BCD per ADC/SBC. Le operazioni binarie
// passano dalla ALU a gate di retronet-logic; in decimal mode la correzione BCD
// viene applicata con ulteriori somme/sottrazioni sullo stesso datapath.
package i6502

import (
	"github.com/retronet-labs/retronet-logic/alu"
	"github.com/retronet-labs/retronet-logic/bit"
	"github.com/retronet-labs/retronet-logic/bus"
	"github.com/retronet-labs/retronet-logic/shifter"
)

// Width e' la larghezza dati del 6502.
const Width = 8

// Op identifica le operazioni logiche bit a bit del 6502.
type Op byte

const (
	OpAND Op = iota
	OpORA
	OpEOR
)

// Flags raccoglie i flag prodotti dalle operazioni aritmetico-logiche del 6502.
// Non tutte le istruzioni applicano tutti i campi: ad esempio BIT usa Zero,
// Negative e Overflow, mentre INC/DEC non toccano Carry/Overflow.
type Flags struct {
	Carry    bool
	Zero     bool
	Negative bool
	Overflow bool
}

// ADC esegue A + value + Carry. In decimal mode restituisce il risultato BCD,
// mentre Zero/Negative/Overflow derivano dal risultato binario pre-correzione,
// come sul 6502 NMOS.
func ADC(a, value byte, carryIn bool, decimal bool) (byte, Flags) {
	bin, carry := add8(a, value, bit.FromBool(carryIn))
	flags := Flags{
		Carry:    carry,
		Zero:     bin == 0,
		Negative: bin&0x80 != 0,
		Overflow: (^(a ^ value) & (a ^ bin) & 0x80) != 0,
	}
	if !decimal {
		return bin, flags
	}

	result := bin
	if (a&0x0F)+(value&0x0F)+b2u(carryIn) > 9 {
		result, _ = add8(result, 0x06, bit.Zero)
	}
	if uint16(a)+uint16(value)+uint16(b2u(carryIn)) > 0x99 {
		result, _ = add8(result, 0x60, bit.Zero)
		flags.Carry = true
	} else {
		flags.Carry = false
	}
	return result, flags
}

// SBC esegue A - value - !Carry. Carry in ingresso e uscita vale 1 quando non
// c'e' prestito. In decimal mode Zero/Negative/Overflow derivano dal risultato
// binario pre-correzione, come sul 6502 NMOS.
func SBC(a, value byte, carryIn bool, decimal bool) (byte, Flags) {
	cin := bit.FromBool(carryIn)
	bin, noBorrow := sub8(a, value, cin)
	flags := Flags{
		Carry:    noBorrow,
		Zero:     bin == 0,
		Negative: bin&0x80 != 0,
		Overflow: ((a ^ value) & (a ^ bin) & 0x80) != 0,
	}
	if !decimal {
		return bin, flags
	}

	result := bin
	if int(a&0x0F)-int(b2u(!carryIn)) < int(value&0x0F) {
		result, _ = sub8(result, 0x06, bit.One)
	}
	if !noBorrow {
		result, _ = sub8(result, 0x60, bit.One)
	}
	return result, flags
}

// Compare calcola reg - value e restituisce i flag C/Z/N del confronto 6502.
func Compare(reg, value byte) (byte, Flags) {
	out, noBorrow := sub8(reg, value, bit.One)
	return out, nz(out, noBorrow)
}

// Logic esegue AND/ORA/EOR e restituisce i flag Z/N.
func Logic(op Op, a, value byte) (byte, Flags) {
	av := bus.FromUint(uint64(a), Width)
	bv := bus.FromUint(uint64(value), Width)
	var out bus.Bus
	switch op {
	case OpAND:
		out, _ = alu.Compute(alu.And, av, bv, bit.Zero)
	case OpEOR:
		out, _ = alu.Compute(alu.Xor, av, bv, bit.Zero)
	default:
		out, _ = alu.Compute(alu.Or, av, bv, bit.Zero)
	}
	v := byte(out.Uint())
	return v, nz(v, false)
}

// BIT restituisce i flag prodotti dall'istruzione BIT: Z da A&value, N/V dai
// bit 7 e 6 dell'operando testato.
func BIT(a, value byte) Flags {
	_, f := Logic(OpAND, a, value)
	return Flags{Zero: f.Zero, Negative: value&0x80 != 0, Overflow: value&0x40 != 0}
}

// Increment esegue value+1 e produce Z/N.
func Increment(value byte) (byte, Flags) {
	out, _ := add8(value, 1, bit.Zero)
	return out, nz(out, false)
}

// Decrement esegue value-1 e produce Z/N.
func Decrement(value byte) (byte, Flags) {
	out, _ := sub8(value, 1, bit.One)
	return out, nz(out, false)
}

// ShiftLeft esegue ASL.
func ShiftLeft(value byte) (byte, Flags) {
	out, carry := shifter.ShiftLeft(bus.FromUint(uint64(value), Width))
	v := byte(out.Uint())
	return v, nz(v, carry.IsHigh())
}

// ShiftRight esegue LSR.
func ShiftRight(value byte) (byte, Flags) {
	out, carry := shifter.ShiftRight(bus.FromUint(uint64(value), Width))
	v := byte(out.Uint())
	return v, nz(v, carry.IsHigh())
}

// RotateLeft esegue ROL.
func RotateLeft(value byte, carryIn bool) (byte, Flags) {
	out, carry := shifter.RotateLeftThroughCarry(bus.FromUint(uint64(value), Width), bit.FromBool(carryIn))
	v := byte(out.Uint())
	return v, nz(v, carry.IsHigh())
}

// RotateRight esegue ROR.
func RotateRight(value byte, carryIn bool) (byte, Flags) {
	out, carry := shifter.RotateRightThroughCarry(bus.FromUint(uint64(value), Width), bit.FromBool(carryIn))
	v := byte(out.Uint())
	return v, nz(v, carry.IsHigh())
}

func add8(a, b byte, cin bit.Bit) (byte, bool) {
	out, f := alu.Compute(alu.Add, bus.FromUint(uint64(a), Width), bus.FromUint(uint64(b), Width), cin)
	return byte(out.Uint()), f.Carry.IsHigh()
}

func sub8(a, b byte, cin bit.Bit) (byte, bool) {
	out, f := alu.Compute(alu.Sub, bus.FromUint(uint64(a), Width), bus.FromUint(uint64(b), Width), cin)
	return byte(out.Uint()), f.Carry.IsHigh()
}

func nz(v byte, carry bool) Flags {
	return Flags{Carry: carry, Zero: v == 0, Negative: v&0x80 != 0}
}

func b2u(v bool) byte {
	if v {
		return 1
	}
	return 0
}
