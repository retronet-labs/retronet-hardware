// Package registerfile implementa un banco di registri (register file): un
// insieme di registri della stessa larghezza con porte di lettura combinatorie e
// una porta di scrittura selezionabile, sincronizzata dal clock.
//
// È costruito sui registri del pacchetto
// [github.com/retronet-labs/retronet-hardware/register]: tutti condividono il
// clock; a ogni fronte di salita scrive solo il registro selezionato (gli altri
// mantengono il valore). La selezione dell'indirizzo è logica di controllo
// (in Go); i registri — il percorso dati — restano costruiti dai gate.
package registerfile

import (
	"github.com/retronet-labs/retronet-logic/bit"
	"github.com/retronet-labs/retronet-logic/bus"

	"github.com/retronet-labs/retronet-hardware/register"
)

// File è un banco di registri.
type File struct {
	regs  []*register.Register
	width int
}

// New crea un register file di count registri, ciascuno di width bit, a zero.
func New(count, width int) *File {
	regs := make([]*register.Register, count)
	for i := range regs {
		regs[i] = register.New(width)
	}
	return &File{regs: regs, width: width}
}

// Count restituisce il numero di registri.
func (f *File) Count() int { return len(f.regs) }

// Width restituisce la larghezza in bit di ogni registro.
func (f *File) Width() int { return f.width }

// Read restituisce il contenuto del registro sel (porta di lettura
// combinatoria, senza clock). È pensata per essere chiamata più volte per
// leggere più operandi nello stesso ciclo.
func (f *File) Read(sel int) bus.Bus {
	return f.regs[sel].Value()
}

// Step applica un fronte di clock a tutto il banco: se write è alto, il registro
// sel cattura data; tutti gli altri mantengono il valore. Un ciclo si simula
// chiamando Step con clk basso e poi alto.
func (f *File) Step(sel int, data bus.Bus, write, clk bit.Bit) {
	for i, r := range f.regs {
		// Abilitazione "one-hot": solo il registro selezionato, e solo se write.
		load := bit.Zero
		if i == sel && write.IsHigh() {
			load = bit.One
		}
		r.Step(data, load, clk)
	}
}
