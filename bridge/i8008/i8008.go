// Package i8008 adatta la ALU a porte di RetroNet Logic alla semantica del
// gruppo ALU dell'Intel 8008.
//
// È il "ponte" che permette all'emulatore 8008 di delegare le operazioni
// aritmetico-logiche alla ALU costruita dai gate (pacchetto
// [github.com/retronet-labs/retronet-logic/alu]): converte gli operandi
// byte ⇄ bus.Bus, sceglie l'operazione e rimappa i flag nella convenzione 8008.
//
// Differenza chiave di convenzione: l'8008 usa il Carry come *borrow* nelle
// sottrazioni (Carry = 1 sul prestito), mentre la ALU di Logic produce
// Carry = 1 per "nessun prestito". L'adattatore inverte il flag dove serve.
package i8008

import (
	"github.com/retronet-labs/retronet-logic/alu"
	"github.com/retronet-labs/retronet-logic/bit"
	"github.com/retronet-labs/retronet-logic/bus"
)

// Width è la larghezza dei dati dell'8008.
const Width = 8

// Codici del gruppo ALU dell'8008 (bit 3-5 dell'opcode).
const (
	GroupADD = 0 // A + value
	GroupADC = 1 // A + value + carry
	GroupSUB = 2 // A - value
	GroupSBB = 3 // A - value - borrow
	GroupAND = 4 // A AND value
	GroupXOR = 5 // A XOR value
	GroupOR  = 6 // A OR value
	GroupCMP = 7 // come SUB, ma si tengono solo i flag
)

// Flags rispecchia i flag di stato dell'Intel 8008.
type Flags struct {
	Carry  bool
	Zero   bool
	Sign   bool
	Parity bool
}

// ALU esegue l'operazione del gruppo (0-7) sull'accumulatore a e sull'operando
// value, dato il carry corrente, usando la ALU a porte di RetroNet Logic.
//
// Restituisce il risultato e i flag con la convenzione 8008. Per GroupCMP il
// risultato va ignorato dal chiamante (l'8008 aggiorna solo i flag).
func ALU(group byte, a, value byte, carryIn bool) (result byte, flags Flags) {
	av := bus.FromUint(uint64(a), Width)
	bv := bus.FromUint(uint64(value), Width)
	cin := bit.FromBool(carryIn)

	switch group & 0x07 {
	case GroupADD:
		out, f := alu.Compute(alu.Add, av, bv, bit.Zero)
		return arith(out, f, false)
	case GroupADC:
		out, f := alu.Compute(alu.Add, av, bv, cin)
		return arith(out, f, false)
	case GroupSUB:
		out, f := alu.Compute(alu.Sub, av, bv, bit.One) // cin=1: sottrazione semplice
		return arith(out, f, true)
	case GroupSBB:
		// SBB: A - value - borrow, con borrow = carry corrente.
		// In termini di sommatore: cin = NOT(borrow).
		out, f := alu.Compute(alu.Sub, av, bv, bit.FromBool(!carryIn))
		return arith(out, f, true)
	case GroupAND:
		out, f := alu.Compute(alu.And, av, bv, bit.Zero)
		return logic(out, f)
	case GroupXOR:
		out, f := alu.Compute(alu.Xor, av, bv, bit.Zero)
		return logic(out, f)
	case GroupOR:
		out, f := alu.Compute(alu.Or, av, bv, bit.Zero)
		return logic(out, f)
	default: // GroupCMP
		out, f := alu.Compute(alu.Sub, av, bv, bit.One)
		return arith(out, f, true)
	}
}

// Increment esegue value + 1 sulla ALU a porte e restituisce il risultato con i
// flag Zero/Sign/Parity. Modella l'istruzione INR dell'8008, che aggiorna questi
// tre flag ma NON tocca il Carry.
func Increment(value byte) (result byte, zero, sign, parity bool) {
	out, f := alu.Compute(alu.Add, bus.FromUint(uint64(value), Width), bus.FromUint(1, Width), bit.Zero)
	return byte(out.Uint()), f.Zero.IsHigh(), f.Sign.IsHigh(), f.Parity.IsHigh()
}

// Decrement esegue value - 1 sulla ALU a porte e restituisce il risultato con i
// flag Zero/Sign/Parity. Modella l'istruzione DCR dell'8008, che aggiorna questi
// tre flag ma NON tocca il Carry.
func Decrement(value byte) (result byte, zero, sign, parity bool) {
	out, f := alu.Compute(alu.Sub, bus.FromUint(uint64(value), Width), bus.FromUint(1, Width), bit.One)
	return byte(out.Uint()), f.Zero.IsHigh(), f.Sign.IsHigh(), f.Parity.IsHigh()
}

// arith costruisce risultato e flag per le operazioni aritmetiche. Per la
// sottrazione (isSub) il Carry dell'8008 è il prestito, cioè NOT del carry ALU.
func arith(out bus.Bus, f alu.Flags, isSub bool) (byte, Flags) {
	carry := f.Carry.IsHigh()
	if isSub {
		carry = !carry
	}
	return byte(out.Uint()), Flags{
		Carry:  carry,
		Zero:   f.Zero.IsHigh(),
		Sign:   f.Sign.IsHigh(),
		Parity: f.Parity.IsHigh(),
	}
}

// logic costruisce risultato e flag per le operazioni logiche, dove l'8008
// azzera sempre il Carry.
func logic(out bus.Bus, f alu.Flags) (byte, Flags) {
	return byte(out.Uint()), Flags{
		Carry:  false,
		Zero:   f.Zero.IsHigh(),
		Sign:   f.Sign.IsHigh(),
		Parity: f.Parity.IsHigh(),
	}
}
