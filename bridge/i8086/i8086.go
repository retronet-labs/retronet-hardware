// Package i8086 adatta la ALU a porte di RetroNet Logic alla semantica
// aritmetico-logica dell'Intel 8086/8088.
//
// Come per i bridge 4004/8008/8080, l'aritmetica non passa per gli operatori di
// Go: addizione e sottrazione (in complemento a due) si appoggiano al sommatore
// a gate di retronet-logic, e i flag derivano dai primitivi logici. L'8086
// aggiunge rispetto all'8080 due novita' modellate qui:
//
//   - larghezza dati selezionabile (8 o 16 bit): la stessa alu.Compute opera su
//     bus.Bus di larghezza qualsiasi, quindi basta scegliere width;
//   - il flag Overflow (OF), assente sull'8080: vale 1 quando il riporto entrante
//     nel bit di segno differisce dal riporto uscente.
//
// Convenzioni di flag dell'8086 riprodotte:
//   - SUB/SBB/CMP usano il Carry come *prestito* (borrow): Carry = NOT(carry-out
//     dell'addizione equivalente a + NOT(b) + cin);
//   - l'Auxiliary Carry e' il riporto del bit 3 (mezzo-byte), senza inversione;
//   - la Parity guarda solo gli 8 bit bassi del risultato, anche a 16 bit.
package i8086

import (
	"github.com/retronet-labs/retronet-logic/alu"
	"github.com/retronet-labs/retronet-logic/bit"
	"github.com/retronet-labs/retronet-logic/bus"
	"github.com/retronet-labs/retronet-logic/gates"
)

// Larghezze dati supportate dall'8086.
const (
	Width8  = 8
	Width16 = 16
)

// Gruppi aritmetico-logici, nello stesso ordine del campo reg del ModR/M nei
// gruppi opcode 0x80-0x83 e del blocco 0x00-0x3F dell'8086.
const (
	GroupADD = 0
	GroupOR  = 1
	GroupADC = 2
	GroupSBB = 3
	GroupAND = 4
	GroupSUB = 5
	GroupXOR = 6
	GroupCMP = 7
)

// Flags rispecchia i sei flag aritmetici dell'8086 prodotti dalla ALU. I flag di
// controllo (Trap, Interrupt, Direction) non nascono dall'aritmetica e vivono
// nella CPU.
type Flags struct {
	Carry     bool
	Parity    bool
	Auxiliary bool
	Zero      bool
	Sign      bool
	Overflow  bool
}

// ALU esegue uno degli otto gruppi aritmetico-logici dell'8086 su a e b alla
// larghezza width (8 o 16), con il carry entrante carryIn. Per CMP il chiamante
// scarta il risultato e tiene solo i flag.
func ALU(group byte, a, b uint16, width int, carryIn bool) (result uint16, flags Flags) {
	switch group & 0x07 {
	case GroupADD:
		return arith(a, b, width, bit.Zero, false)
	case GroupADC:
		return arith(a, b, width, bit.FromBool(carryIn), false)
	case GroupSUB, GroupCMP:
		return arith(a, b, width, bit.One, true)
	case GroupSBB:
		return arith(a, b, width, bit.FromBool(!carryIn), true)
	case GroupAND:
		return logic(alu.And, a, b, width)
	case GroupOR:
		return logic(alu.Or, a, b, width)
	default: // GroupXOR
		return logic(alu.Xor, a, b, width)
	}
}

// Increment esegue value + 1: aggiorna OF/SF/ZF/AF/PF ma NON il Carry (come INC
// sull'8086). Il chiamante conserva il proprio Carry.
func Increment(value uint16, width int) (uint16, Flags) {
	out, f := arith(value, 1, width, bit.Zero, false)
	f.Carry = false
	return out, f
}

// Decrement esegue value - 1: aggiorna OF/SF/ZF/AF/PF ma NON il Carry (come DEC).
func Decrement(value uint16, width int) (uint16, Flags) {
	out, f := arith(value, 1, width, bit.One, true)
	f.Carry = false
	return out, f
}

// Mul moltiplica a per b (width bit ciascuno) col metodo shift-and-add sul
// sommatore a gate, restituendo il prodotto a 2*width bit. Con signed=true tratta
// gli operandi come interi con segno (moltiplica i moduli e applica il segno).
// overflow segnala se la meta' alta del prodotto e' significativa: e' la
// condizione che l'8086 usa per CF/OF di MUL (meta' alta != 0) e IMUL (meta' alta
// != estensione di segno della meta' bassa).
func Mul(a, b uint16, width int, signed bool) (product uint32, overflow bool) {
	w2 := width * 2
	av := uint64(a) & widthMask(width)
	bv := uint64(b) & widthMask(width)

	var prod uint64
	if signed {
		magA, negA := absWidth(av, width)
		magB, negB := absWidth(bv, width)
		prod = mulMagnitude(magA, magB, width)
		if negA != negB {
			prod = negate(prod, w2)
		}
	} else {
		prod = mulMagnitude(av, bv, width)
	}
	prod &= widthMask(w2)

	hi := prod >> uint(width)
	lo := prod & widthMask(width)
	if signed {
		if msb(lo, width) {
			overflow = hi != widthMask(width) // attesa: estensione di segno tutta a 1
		} else {
			overflow = hi != 0
		}
	} else {
		overflow = hi != 0
	}
	return uint32(prod), overflow
}

// Div divide il dividendo (2*width bit) per divisor (width bit) con divisione a
// ripristino sul sommatore a gate. Restituisce quoziente e resto (width bit) e
// ok=false in caso di errore di divisione dell'8086: divisore nullo oppure
// quoziente fuori intervallo. Con signed=true esegue la divisione con segno
// (troncata verso zero, resto col segno del dividendo).
func Div(dividend uint32, divisor uint16, width int, signed bool) (quot, rem uint16, ok bool) {
	w2 := width * 2
	dv := uint64(divisor) & widthMask(width)
	dd := uint64(dividend) & widthMask(w2)
	if dv == 0 {
		return 0, 0, false
	}

	if !signed {
		q, r, fit := divMagnitude(dd, dv, width)
		if !fit {
			return 0, 0, false
		}
		return uint16(q), uint16(r), true
	}

	magDividend, negDividend := absWidth(dd, w2)
	magDivisor, negDivisor := absWidth(dv, width)
	qMag, rMag, _ := divMagnitude(magDividend, magDivisor, width)

	// Controllo di intervallo sul MODULO, prima di troncare a width: il quoziente
	// con segno deve stare in [-2^(w-1), 2^(w-1)-1].
	resultNeg := negDividend != negDivisor
	var limit uint64
	if resultNeg {
		limit = uint64(1) << uint(width-1) // ammesso -2^(w-1)
	} else {
		limit = (uint64(1) << uint(width-1)) - 1
	}
	if qMag > limit {
		return 0, 0, false
	}

	q := qMag
	if resultNeg {
		q = negate(qMag, width)
	}
	r := rMag
	if negDividend {
		r = negate(rMag, width) // il resto prende il segno del dividendo
	}
	return uint16(q & widthMask(width)), uint16(r & widthMask(width)), true
}

// mulMagnitude moltiplica due moduli unsigned con shift-and-add sul sommatore a
// gate, accumulando su 2*width bit.
func mulMagnitude(a, b uint64, width int) uint64 {
	w2 := width * 2
	product := uint64(0)
	for i := 0; i < width; i++ {
		if (b>>uint(i))&1 == 1 {
			product, _ = add(product, (a<<uint(i))&widthMask(w2), w2, bit.Zero)
		}
	}
	return product & widthMask(w2)
}

// divMagnitude esegue la divisione unsigned a ripristino: per ogni bit del
// dividendo (dal piu' significativo) sposta il resto, prova a sottrarre il
// divisore con il sommatore a gate e, se non c'e' prestito, fissa il bit del
// quoziente. fit=false se il quoziente non entra in width bit.
func divMagnitude(dividend, divisor uint64, width int) (quot, rem uint64, fit bool) {
	w2 := width * 2
	for i := w2 - 1; i >= 0; i-- {
		rem = (rem << 1) | (dividend>>uint(i))&1
		diff, noBorrow := sub(rem, divisor, w2)
		if noBorrow {
			rem = diff
			quot |= uint64(1) << uint(i)
		}
	}
	return quot & widthMask(w2), rem & widthMask(width), quot <= widthMask(width)
}

// sub calcola a - b a gate restituendo (risultato, noBorrow); noBorrow=true
// significa a >= b (nessun prestito).
func sub(a, b uint64, width int) (uint64, bool) {
	out, carryOut := add(a, (^b)&widthMask(width), width, bit.One)
	return out, carryOut
}

// absWidth restituisce il modulo di x interpretato con segno entro width e se era
// negativo; la negazione e' il complemento a due calcolato a gate.
func absWidth(x uint64, width int) (mag uint64, neg bool) {
	if msb(x, width) {
		return negate(x, width), true
	}
	return x & widthMask(width), false
}

// negate calcola il complemento a due (0 - x) entro width sul sommatore a gate.
func negate(x uint64, width int) uint64 {
	out, _ := add(0, (^x)&widthMask(width), width, bit.One)
	return out & widthMask(width)
}

// arith esegue l'addizione a + addend + cin, dove per la sottrazione addend e'
// il complemento di b entro width e isSub inverte il Carry (borrow). Tutti i
// flag derivano dal sommatore e dai gate, senza operatori aritmetici di Go.
func arith(a, b uint16, width int, cin bit.Bit, isSub bool) (uint16, Flags) {
	mask := widthMask(width)
	av := uint64(a) & mask
	bv := uint64(b) & mask
	addend := bv
	if isSub {
		addend = (^bv) & mask // complemento a uno entro width; il +1 lo porta cin
	}

	out, carryOut := add(av, addend, width, cin)
	carryMSB := carryInto(av, addend, width-1, cin) // riporto entrante nel bit di segno
	carry4 := carryInto(av, addend, 4, cin)         // riporto del mezzo-byte (AF)

	carry := carryOut
	if isSub {
		carry = !carryOut // sull'8086 il Carry della sottrazione e' il prestito
	}

	return uint16(out), Flags{
		Carry:     carry,
		Parity:    parityEvenLow8(out),
		Auxiliary: carry4,
		Zero:      out == 0,
		Sign:      msb(out, width),
		Overflow:  carryMSB != carryOut,
	}
}

// logic esegue un'operazione logica bit a bit: Carry e Overflow azzerati,
// Auxiliary indefinito sull'8086 (qui 0, coerente col backend native).
func logic(op alu.Op, a, b uint16, width int) (uint16, Flags) {
	mask := widthMask(width)
	out, _ := alu.Compute(op, bus.FromUint(uint64(a)&mask, width), bus.FromUint(uint64(b)&mask, width), bit.Zero)
	v := out.Uint()
	return uint16(v), Flags{
		Parity: parityEvenLow8(v),
		Zero:   v == 0,
		Sign:   msb(v, width),
	}
}

// add somma a + b + cin a gate, restituendo risultato e riporto uscente.
func add(a, b uint64, width int, cin bit.Bit) (uint64, bool) {
	out, f := alu.Compute(alu.Add, bus.FromUint(a, width), bus.FromUint(b, width), cin)
	return out.Uint(), f.Carry.IsHigh()
}

// carryInto restituisce il riporto entrante nel bit bitNo dell'addizione
// a + b + cin, ottenuto sommando a gate solo i bitNo bit meno significativi.
func carryInto(a, b uint64, bitNo int, cin bit.Bit) bool {
	if bitNo <= 0 {
		return cin.IsHigh()
	}
	lowMask := (uint64(1) << uint(bitNo)) - 1
	_, f := alu.Compute(alu.Add, bus.FromUint(a&lowMask, bitNo), bus.FromUint(b&lowMask, bitNo), cin)
	return f.Carry.IsHigh()
}

// parityEvenLow8 vale true se il numero di bit a 1 negli 8 bit bassi e' pari:
// riduzione XOR a gate, poi negata (parita' pari).
func parityEvenLow8(v uint64) bool {
	acc := bit.Zero
	for _, x := range bus.FromUint(v&0xFF, 8) {
		acc = gates.Xor(acc, x)
	}
	return gates.Not(acc).IsHigh()
}

func msb(v uint64, width int) bool {
	return (v>>uint(width-1))&1 == 1
}

func widthMask(width int) uint64 {
	return (uint64(1) << uint(width)) - 1
}
