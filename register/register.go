// Package register implementa un registro a N bit: il primo componente capace
// di memorizzare una parola intera.
//
// Un registro è semplicemente un insieme di N D flip-flop
// ([github.com/retronet-labs/retronet-hardware/flipflop]) che condividono lo
// stesso clock, con un ingresso di abilitazione (load) che decide, a ogni
// fronte, se caricare un nuovo dato o mantenere quello presente.
package register

import (
	"github.com/retronet-labs/retronet-logic/bit"
	"github.com/retronet-labs/retronet-logic/bus"
	"github.com/retronet-labs/retronet-logic/gates"

	"github.com/retronet-labs/retronet-hardware/flipflop"
)

// Register è un registro a N bit.
type Register struct {
	bits []*flipflop.DFlipFlop
}

// New crea un registro di width bit, inizializzato a zero.
func New(width int) *Register {
	ffs := make([]*flipflop.DFlipFlop, width)
	for i := range ffs {
		ffs[i] = flipflop.NewDFlipFlop()
	}
	return &Register{bits: ffs}
}

// Width restituisce il numero di bit del registro.
func (r *Register) Width() int {
	return len(r.bits)
}

// Step valuta il registro al livello di clock indicato e restituisce il
// contenuto. Se load = 1, al fronte di salita carica data; se load = 0, mantiene
// il valore corrente. Un ciclo si simula chiamando Step con clk basso e poi
// alto.
//
// Va in panico se la larghezza di data è diversa da quella del registro.
func (r *Register) Step(data bus.Bus, load, clk bit.Bit) bus.Bus {
	if data.Width() != len(r.bits) {
		panic("register: larghezza del dato diversa da quella del registro")
	}
	out := make(bus.Bus, len(r.bits))
	for i, ff := range r.bits {
		// MUX 2:1 per bit: con load si sceglie il nuovo dato, altrimenti si
		// reimmette il valore attuale (mantenimento).
		dEff := mux1(load, ff.Q(), data[i])
		out[i] = ff.Step(dEff, clk)
	}
	return out
}

// Value restituisce il contenuto corrente del registro senza modificarlo.
func (r *Register) Value() bus.Bus {
	out := make(bus.Bus, len(r.bits))
	for i, ff := range r.bits {
		out[i] = ff.Q()
	}
	return out
}

// mux1 è un multiplexer 2:1 a un bit: restituisce b se sel = 1, altrimenti a.
//
//	mux1 = OR(AND(a, NOT sel), AND(b, sel))
func mux1(sel, a, b bit.Bit) bit.Bit {
	return gates.Or(gates.And(a, gates.Not(sel)), gates.And(b, sel))
}
