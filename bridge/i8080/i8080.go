// Package i8080 adatta la ALU a porte di RetroNet Logic alla semantica
// aritmetico-logica dell'Intel 8080.
package i8080

import (
	"github.com/retronet-labs/retronet-logic/alu"
	"github.com/retronet-labs/retronet-logic/bit"
	"github.com/retronet-labs/retronet-logic/bus"
)

// Width e' la larghezza dati dell'Intel 8080.
const Width = 8

const (
	GroupADD = 0
	GroupADC = 1
	GroupSUB = 2
	GroupSBB = 3
	GroupANA = 4
	GroupXRA = 5
	GroupORA = 6
	GroupCMP = 7
)

// Flags rispecchia i flag aritmetici dell'Intel 8080.
type Flags struct {
	Carry          bool
	Zero           bool
	Sign           bool
	Parity         bool
	AuxiliaryCarry bool
}

// ALU esegue uno dei gruppi aritmetico-logici 8080 su A e value.
func ALU(group byte, a, value byte, carryIn bool) (result byte, flags Flags) {
	av := bus.FromUint(uint64(a), Width)
	bv := bus.FromUint(uint64(value), Width)
	cin := bit.FromBool(carryIn)

	switch group & 0x07 {
	case GroupADD:
		out, f := alu.Compute(alu.Add, av, bv, bit.Zero)
		return arith(out, f, auxAdd(a, value, false), false)
	case GroupADC:
		out, f := alu.Compute(alu.Add, av, bv, cin)
		return arith(out, f, auxAdd(a, value, carryIn), false)
	case GroupSUB:
		out, f := alu.Compute(alu.Sub, av, bv, bit.One)
		return arith(out, f, auxSub(a, value, false), true)
	case GroupSBB:
		out, f := alu.Compute(alu.Sub, av, bv, bit.FromBool(!carryIn))
		return arith(out, f, auxSub(a, value, carryIn), true)
	case GroupANA:
		out, f := alu.Compute(alu.And, av, bv, bit.Zero)
		return logic(out, f, auxAnd(a, value))
	case GroupXRA:
		out, f := alu.Compute(alu.Xor, av, bv, bit.Zero)
		return logic(out, f, false)
	case GroupORA:
		out, f := alu.Compute(alu.Or, av, bv, bit.Zero)
		return logic(out, f, false)
	default:
		out, f := alu.Compute(alu.Sub, av, bv, bit.One)
		return arith(out, f, auxSub(a, value, false), true)
	}
}

// Increment esegue value + 1 e restituisce i flag aggiornati dall'istruzione INR.
func Increment(value byte) (result byte, flags Flags) {
	result, flags = ALU(GroupADD, value, 1, false)
	flags.Carry = false
	return result, flags
}

// Decrement esegue value - 1 e restituisce i flag aggiornati dall'istruzione DCR.
func Decrement(value byte) (result byte, flags Flags) {
	result, flags = ALU(GroupSUB, value, 1, false)
	flags.Carry = false
	return result, flags
}

// Add16 somma due parole a 16 bit usando due ALU a 8 bit in cascata.
func Add16(a, value uint16) (result uint16, carry bool) {
	low, lowFlags := ALU(GroupADD, byte(a), byte(value), false)
	high, highFlags := ALU(GroupADC, byte(a>>8), byte(value>>8), lowFlags.Carry)
	return uint16(high)<<8 | uint16(low), highFlags.Carry
}

func arith(out bus.Bus, f alu.Flags, aux bool, isSub bool) (byte, Flags) {
	carry := f.Carry.IsHigh()
	if isSub {
		carry = !carry
	}
	return byte(out.Uint()), Flags{
		Carry:          carry,
		Zero:           f.Zero.IsHigh(),
		Sign:           f.Sign.IsHigh(),
		Parity:         f.Parity.IsHigh(),
		AuxiliaryCarry: aux,
	}
}

func logic(out bus.Bus, f alu.Flags, aux bool) (byte, Flags) {
	return byte(out.Uint()), Flags{
		Zero:           f.Zero.IsHigh(),
		Sign:           f.Sign.IsHigh(),
		Parity:         f.Parity.IsHigh(),
		AuxiliaryCarry: aux,
	}
}

func auxAdd(a, value byte, carryIn bool) bool {
	_, f := alu.Compute(alu.Add, lowNibble(a), lowNibble(value), bit.FromBool(carryIn))
	return f.Carry.IsHigh()
}

// auxSub calcola l'Auxiliary Carry (half-borrow) di SUB/SBB/CMP. Sull'8080 è il
// carry-out del bit 3 dell'addizione equivalente a + NOT(value) + (!borrowIn),
// senza negazione (a differenza del Carry pieno, che invece è il prestito).
func auxSub(a, value byte, borrowIn bool) bool {
	_, f := alu.Compute(alu.Sub, lowNibble(a), lowNibble(value), bit.FromBool(!borrowIn))
	return f.Carry.IsHigh()
}

// auxAnd calcola l'Auxiliary Carry dell'istruzione ANA dell'8080: per quirk
// dell'hardware è il bit 3 di (A OR value), calcolato qui con la OR a gate.
func auxAnd(a, value byte) bool {
	out, _ := alu.Compute(alu.Or, bus.FromUint(uint64(a), Width), bus.FromUint(uint64(value), Width), bit.Zero)
	return out[3].IsHigh()
}

func lowNibble(v byte) bus.Bus {
	return bus.FromUint(uint64(v&0x0F), 4)
}
