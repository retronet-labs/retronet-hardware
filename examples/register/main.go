// Command register dimostra un registro a 8 bit: caricamento, mantenimento e
// nuovo caricamento, scanditi dal clock.
//
// Esecuzione:
//
//	go run ./examples/register
package main

import (
	"fmt"

	"github.com/retronet-labs/retronet-logic/bit"
	"github.com/retronet-labs/retronet-logic/bus"

	"github.com/retronet-labs/retronet-hardware/register"
)

// ciclo esegue un colpo di clock completo (fase bassa + fronte di salita).
func ciclo(r *register.Register, valore uint64, load bit.Bit) {
	data := bus.FromUint(valore, r.Width())
	r.Step(data, load, bit.Zero)
	r.Step(data, load, bit.One)
}

func main() {
	r := register.New(8)
	fmt.Printf("iniziale       -> %s (%d)\n", r.Value(), r.Value().Uint())

	ciclo(r, 42, bit.One)
	fmt.Printf("load 42        -> %s (%d)\n", r.Value(), r.Value().Uint())

	ciclo(r, 99, bit.Zero)
	fmt.Printf("99 con load=0  -> %s (%d)  [mantiene]\n", r.Value(), r.Value().Uint())

	ciclo(r, 99, bit.One)
	fmt.Printf("load 99        -> %s (%d)\n", r.Value(), r.Value().Uint())
}
