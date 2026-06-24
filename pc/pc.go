// Package pc implementa il Program Counter: il registro che tiene l'indirizzo
// della prossima istruzione e sa incrementarsi o caricare un nuovo indirizzo
// (per i salti).
//
// È un esempio di componente sequenziale costruito combinando uno stato (un
// [register]) con logica combinatoria di RetroNet Logic: l'incremento usa il
// sommatore [adder], la scelta del prossimo valore usa i [mux] e le [gates].
package pc

import (
	"github.com/retronet-labs/retronet-logic/adder"
	"github.com/retronet-labs/retronet-logic/bit"
	"github.com/retronet-labs/retronet-logic/bus"
	"github.com/retronet-labs/retronet-logic/gates"
	"github.com/retronet-labs/retronet-logic/mux"

	"github.com/retronet-labs/retronet-hardware/register"
)

// PC è un program counter a width bit.
type PC struct {
	reg   *register.Register
	width int
}

// New crea un Program Counter di width bit, inizializzato a 0.
func New(width int) *PC {
	return &PC{reg: register.New(width), width: width}
}

// Value restituisce l'indirizzo corrente (combinatorio, senza clock).
func (p *PC) Value() bus.Bus {
	return p.reg.Value()
}

// Step aggiorna il PC al fronte di salita del clock secondo le priorità:
//
//   - se load = 1: PC = addr (salto);
//   - altrimenti se inc = 1: PC = PC + 1;
//   - altrimenti: mantiene il valore.
//
// addr deve avere la stessa larghezza del PC. Un ciclo si simula chiamando Step
// con clk basso e poi alto.
func (p *PC) Step(addr bus.Bus, load, inc, clk bit.Bit) {
	cur := p.reg.Value()
	incremented, _ := adder.Add(cur, bus.FromUint(1, p.width), bit.Zero)

	// next = load ? addr : (inc ? incremented : cur)
	next := mux.TwoBus(load, mux.TwoBus(inc, cur, incremented), addr)
	write := gates.Or(load, inc)
	p.reg.Step(next, write, clk)
}
