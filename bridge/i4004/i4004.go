// Package i4004 adatta la ALU a porte di RetroNet Logic alle operazioni
// aritmetiche dell'Intel 4004 (processore a 4 bit).
//
// È il "ponte" che permette all'emulatore 4004 di delegare le sue operazioni
// aritmetiche alla ALU costruita dai gate (pacchetto
// [github.com/retronet-labs/retronet-logic/alu]).
//
// A differenza dell'8008, il 4004 usa il Carry nella stessa convenzione della
// ALU di Logic (carryIn incluso direttamente, carryOut = riporto del sommatore a
// 4 bit): non serve alcuna inversione. Il 4004 ha un unico flag, il Carry.
package i4004

import (
	"github.com/retronet-labs/retronet-logic/alu"
	"github.com/retronet-labs/retronet-logic/bit"
	"github.com/retronet-labs/retronet-logic/bus"
)

// Width è la larghezza dei dati del 4004 (nibble).
const Width = 4

// nib costruisce un bus a 4 bit dal nibble basso di v.
func nib(v byte) bus.Bus {
	return bus.FromUint(uint64(v), Width)
}

// Add esegue ADD: a + r + carryIn. Restituisce il nibble risultante e il carry.
func Add(a, r byte, carryIn bool) (result byte, carry bool) {
	out, f := alu.Compute(alu.Add, nib(a), nib(r), bit.FromBool(carryIn))
	return byte(out.Uint()), f.Carry.IsHigh()
}

// Sub esegue SUB: a + NOT(r) + carryIn (sottrazione in complemento a due come
// sul 4004). Restituisce il nibble risultante e il carry (1 = nessun prestito).
func Sub(a, r byte, carryIn bool) (result byte, carry bool) {
	out, f := alu.Compute(alu.Sub, nib(a), nib(r), bit.FromBool(carryIn))
	return byte(out.Uint()), f.Carry.IsHigh()
}

// Inc esegue IAC (Increment Accumulator): a + 1.
func Inc(a byte) (result byte, carry bool) {
	out, f := alu.Compute(alu.Add, nib(a), nib(1), bit.Zero)
	return byte(out.Uint()), f.Carry.IsHigh()
}

// Dec esegue DAC (Decrement Accumulator): a - 1, realizzato come a + NOT(1) + 1
// (equivalente all'a + 0x0F dell'hardware). carry = 1 indica nessun prestito.
func Dec(a byte) (result byte, carry bool) {
	out, f := alu.Compute(alu.Sub, nib(a), nib(1), bit.One)
	return byte(out.Uint()), f.Carry.IsHigh()
}

// Complement esegue CMA (Complement Accumulator): NOT(a), senza toccare il carry.
func Complement(a byte) (result byte) {
	out, _ := alu.Compute(alu.Not, nib(a), nib(0), bit.Zero)
	return byte(out.Uint())
}
