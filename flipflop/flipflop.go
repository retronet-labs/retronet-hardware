// Package flipflop implementa gli elementi di memoria sincronizzati dal clock:
// il D latch (sensibile al livello) e il D flip-flop (sensibile al fronte).
//
// Entrambi si costruiscono a partire dal latch SR del pacchetto
// [github.com/retronet-labs/retronet-hardware/latch], aggiungendo la logica di
// gating con il clock realizzata con le porte di RetroNet Logic.
package flipflop

import (
	"github.com/retronet-labs/retronet-logic/bit"
	"github.com/retronet-labs/retronet-logic/gates"

	"github.com/retronet-labs/retronet-hardware/latch"
)

// DLatch è un latch D ("gated"): elimina lo stato non valido del latch SR
// derivando S e R da un unico ingresso dato (D) abilitato dal clock.
//
//	S = AND(D, clk)
//	R = AND(NOT D, clk)
//
// Quando clk = 1 è "trasparente" (Q segue D); quando clk = 0 mantiene il valore.
type DLatch struct {
	sr *latch.SR
}

// NewDLatch crea un D latch nello stato di reset (Q = 0).
func NewDLatch() *DLatch {
	return &DLatch{sr: latch.NewSR()}
}

// Step applica il dato d con il clock clk e restituisce Q.
func (d *DLatch) Step(data, clk bit.Bit) bit.Bit {
	s := gates.And(data, clk)
	r := gates.And(gates.Not(data), clk)
	q, _ := d.sr.Step(s, r)
	return q
}

// Q restituisce lo stato corrente senza modificarlo.
func (d *DLatch) Q() bit.Bit {
	return d.sr.Q()
}

// DFlipFlop è un D flip-flop sensibile al fronte di salita del clock, costruito
// con lo schema master-slave: due D latch con clock opposti. Il master è
// trasparente mentre il clock è basso (insegue D); sul fronte di salita il suo
// valore congelato viene trasferito allo slave, che lo presenta in uscita.
//
//	          clk'        clk
//	D ──► [ DLatch ] ──► [ DLatch ] ──► Q
//	        master         slave
//	clk' = NOT clk
type DFlipFlop struct {
	master *DLatch
	slave  *DLatch
}

// NewDFlipFlop crea un D flip-flop nello stato di reset (Q = 0).
func NewDFlipFlop() *DFlipFlop {
	return &DFlipFlop{master: NewDLatch(), slave: NewDLatch()}
}

// Step valuta il flip-flop al livello di clock indicato e restituisce Q.
// Un ciclo si simula chiamando Step con clk basso e poi alto: il dato presente
// al fronte di salita viene catturato.
func (ff *DFlipFlop) Step(data, clk bit.Bit) bit.Bit {
	mq := ff.master.Step(data, gates.Not(clk))
	return ff.slave.Step(mq, clk)
}

// Q restituisce lo stato corrente (uscita dello slave) senza modificarlo.
func (ff *DFlipFlop) Q() bit.Bit {
	return ff.slave.Q()
}
