// Package latch implementa il latch SR (Set-Reset), la più semplice cella di
// memoria e il primo componente *con stato* di RetroNet Hardware.
//
// A differenza dei componenti combinatori, un latch ha retroazione: le sue
// uscite rientrano negli ingressi. In simulazione questo si modella con uno
// stato interno aggiornato a ogni Step, iterando finché si stabilizza.
package latch

import (
	"github.com/retronet-labs/retronet-logic/bit"
	"github.com/retronet-labs/retronet-logic/gates"
)

// stabilizeIter è il numero massimo di passate di valutazione delle NOR a
// retroazione incrociata: la convergenza avviene in pochissimi passi.
const stabilizeIter = 8

// SR è un latch Set-Reset costruito con due porte NOR a retroazione incrociata:
//
//	Q  = NOR(R, Qn)
//	Qn = NOR(S, Q)
//
// Custodisce lo stato (Q, Qn) tra una chiamata e l'altra di Step.
//
//	      ┌──────┐
//	R ────┤ NOR  ├──┬──── Q
//	   ┌──┤      │  │
//	   │  └──────┘  │
//	   │  ┌──────┐  │
//	   └──┤ NOR  ├──┘
//	S ────┤      ├──────── Qn
//	      └──────┘
type SR struct {
	q  bit.Bit
	qn bit.Bit
}

// NewSR crea un latch SR nello stato di reset (Q=0, Qn=1).
func NewSR() *SR {
	return &SR{q: bit.Zero, qn: bit.One}
}

// Step applica gli ingressi Set (s) e Reset (r) e restituisce le uscite Q e Qn,
// aggiornando lo stato interno.
//
// Comportamento:
//
//	S R │ Q       (stato successivo)
//	────┼──────────────────────────
//	0 0 │ Q       (mantiene il valore)
//	1 0 │ 1       (set)
//	0 1 │ 0       (reset)
//	1 1 │ — stato non valido: Q e Qn entrambi 0 (da evitare)
func (l *SR) Step(s, r bit.Bit) (q, qn bit.Bit) {
	// Retroazione incrociata: si rivalutano le due NOR finché lo stato è stabile.
	for i := 0; i < stabilizeIter; i++ {
		nq := gates.Nor(r, l.qn)
		nqn := gates.Nor(s, nq)
		if nq == l.q && nqn == l.qn {
			break
		}
		l.q, l.qn = nq, nqn
	}
	return l.q, l.qn
}

// Q restituisce lo stato corrente memorizzato senza modificarlo.
func (l *SR) Q() bit.Bit {
	return l.q
}
