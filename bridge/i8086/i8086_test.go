package i8086

import (
	"math/bits"
	"testing"
)

// --- Casi puntuali (documentano le convenzioni di flag dell'8086) ---

func TestAddOverflowSignedWrap(t *testing.T) {
	// 0x7F + 0x01 = 0x80: overflow con segno (positivo + positivo -> negativo).
	out, f := ALU(GroupADD, 0x7F, 0x01, Width8, false)
	if out != 0x80 || !f.Overflow || !f.Sign || f.Carry {
		t.Fatalf("ADD 0x7F+1 = 0x%02X flags=%+v", out, f)
	}
}

func TestAddCarryNoOverflow(t *testing.T) {
	// 0xFF + 0x01 = 0x00: riporto senza overflow con segno; zero, AF e PF a 1.
	out, f := ALU(GroupADD, 0xFF, 0x01, Width8, false)
	if out != 0x00 || !f.Carry || f.Overflow || !f.Zero || !f.Auxiliary {
		t.Fatalf("ADD 0xFF+1 = 0x%02X flags=%+v", out, f)
	}
}

func TestSubBorrowConvention(t *testing.T) {
	// 0x00 - 0x01 = 0xFF: prestito (Carry=1) e segno negativo.
	out, f := ALU(GroupSUB, 0x00, 0x01, Width8, false)
	if out != 0xFF || !f.Carry || !f.Sign {
		t.Fatalf("SUB 0-1 = 0x%02X flags=%+v", out, f)
	}
}

func TestSub16(t *testing.T) {
	out, f := ALU(GroupSUB, 0x1234, 0x0234, Width16, false)
	if out != 0x1000 || f.Carry {
		t.Fatalf("SUB16 0x1234-0x0234 = 0x%04X flags=%+v", out, f)
	}
}

func TestLogicClearsCarryAndOverflow(t *testing.T) {
	out, f := ALU(GroupAND, 0xF0, 0x0F, Width8, false)
	if out != 0 || f.Carry || f.Overflow || !f.Zero {
		t.Fatalf("AND = 0x%02X flags=%+v", out, f)
	}
}

func TestIncrementPreservesCarryFlag(t *testing.T) {
	// INC su 0xFF -> 0x00: Carry deve restare false (l'8086 non lo tocca).
	out, f := Increment(0xFF, Width8)
	if out != 0x00 || f.Carry || !f.Zero || !f.Auxiliary {
		t.Fatalf("INC 0xFF = 0x%02X flags=%+v", out, f)
	}
}

// --- Oracolo di riferimento 8086 per il test differenziale ---

func refParityEven(v uint32) bool { return bits.OnesCount32(v&0xFF)%2 == 0 }

func refArith(a, b uint32, width int, cinBool, isSub bool) (uint32, Flags) {
	mask := uint32(1)<<uint(width) - 1
	av := a & mask
	bv := b & mask
	addend := bv
	if isSub {
		addend = (^bv) & mask
	}
	cin := uint32(0)
	if cinBool {
		cin = 1
	}
	res := av + addend + cin
	out := res & mask
	xorc := res ^ av ^ addend
	carryOut := res>>uint(width)&1 == 1
	carryMSB := xorc>>uint(width-1)&1 == 1
	af := (av^bv^out)>>4&1 == 1 // AF dai valori originali (b non complementato)

	carry := carryOut
	if isSub {
		carry = !carryOut
	}
	return out, Flags{
		Carry:     carry,
		Parity:    refParityEven(out),
		Auxiliary: af,
		Zero:      out == 0,
		Sign:      out>>uint(width-1)&1 == 1,
		Overflow:  carryMSB != carryOut,
	}
}

func refLogic(op byte, a, b uint32, width int) (uint32, Flags) {
	mask := uint32(1)<<uint(width) - 1
	av := a & mask
	bv := b & mask
	var out uint32
	switch op {
	case GroupAND:
		out = av & bv
	case GroupOR:
		out = av | bv
	default: // GroupXOR
		out = av ^ bv
	}
	return out, Flags{
		Parity: refParityEven(out),
		Zero:   out == 0,
		Sign:   out>>uint(width-1)&1 == 1,
	}
}

func ref8086(group byte, a, b uint32, width int, cin bool) (uint32, Flags) {
	switch group {
	case GroupADD:
		return refArith(a, b, width, false, false)
	case GroupADC:
		return refArith(a, b, width, cin, false)
	case GroupSUB, GroupCMP:
		return refArith(a, b, width, true, true)
	case GroupSBB:
		return refArith(a, b, width, !cin, true)
	default: // GroupAND, GroupOR, GroupXOR
		return refLogic(group, a, b, width)
	}
}

var allGroups = []byte{GroupADD, GroupOR, GroupADC, GroupSBB, GroupAND, GroupSUB, GroupXOR, GroupCMP}

// TestALUDifferentialVs8086Reference8Bit confronta la ALU a gate con l'oracolo
// su TUTTI gli ingressi a 8 bit, tutti i gruppi e i due valori di carry entrante:
// risultato e ogni flag (CF, PF, AF, ZF, SF, OF).
func TestALUDifferentialVs8086Reference8Bit(t *testing.T) {
	for _, g := range allGroups {
		for a := 0; a <= 0xFF; a++ {
			for b := 0; b <= 0xFF; b++ {
				for _, cin := range []bool{false, true} {
					gotR, gotF := ALU(g, uint16(a), uint16(b), Width8, cin)
					wantR, wantF := ref8086(g, uint32(a), uint32(b), Width8, cin)
					if uint32(gotR) != wantR || gotF != wantF {
						t.Fatalf("g=%d a=%#02x b=%#02x cin=%v: got(%#02x,%+v) want(%#02x,%+v)",
							g, a, b, cin, gotR, gotF, wantR, wantF)
					}
				}
			}
		}
	}
}

// TestALUDifferentialVs8086Reference16Bit campiona lo spazio a 16 bit (passo non
// allineato alle potenze di due) per coprire riporti e overflow senza esplodere.
func TestALUDifferentialVs8086Reference16Bit(t *testing.T) {
	const step = 277 // primo, evita allineamenti regolari
	for _, g := range allGroups {
		for a := 0; a <= 0xFFFF; a += step {
			for b := 0; b <= 0xFFFF; b += step {
				for _, cin := range []bool{false, true} {
					gotR, gotF := ALU(g, uint16(a), uint16(b), Width16, cin)
					wantR, wantF := ref8086(g, uint32(a), uint32(b), Width16, cin)
					if uint32(gotR) != wantR || gotF != wantF {
						t.Fatalf("g=%d a=%#04x b=%#04x cin=%v: got(%#04x,%+v) want(%#04x,%+v)",
							g, a, b, cin, gotR, gotF, wantR, wantF)
					}
				}
			}
		}
	}
}

// --- Moltiplicazione e divisione (composte a gate) vs oracolo Go ---

func signExtend(v uint64, width int) int64 {
	if v>>uint(width-1)&1 == 1 {
		return int64(v) - int64(uint64(1)<<uint(width))
	}
	return int64(v)
}

func refMul(a, b uint16, width int, signed bool) (uint32, bool) {
	mask := uint64(1)<<uint(width) - 1
	w2 := width * 2
	maskw2 := uint64(1)<<uint(w2) - 1
	av := uint64(a) & mask
	bv := uint64(b) & mask
	var prod uint64
	if signed {
		prod = uint64(signExtend(av, width)*signExtend(bv, width)) & maskw2
	} else {
		prod = (av * bv) & maskw2
	}
	hi := prod >> uint(width)
	lo := prod & mask
	var of bool
	switch {
	case !signed:
		of = hi != 0
	case lo>>uint(width-1)&1 == 1:
		of = hi != mask
	default:
		of = hi != 0
	}
	return uint32(prod), of
}

func refDiv(dividend uint32, divisor uint16, width int, signed bool) (uint16, uint16, bool) {
	mask := uint64(1)<<uint(width) - 1
	w2 := width * 2
	maskw2 := uint64(1)<<uint(w2) - 1
	dv := uint64(divisor) & mask
	dd := uint64(dividend) & maskw2
	if dv == 0 {
		return 0, 0, false
	}
	if !signed {
		q := dd / dv
		if q > mask {
			return 0, 0, false
		}
		return uint16(q), uint16(dd % dv), true
	}
	sdd := signExtend(dd, w2)
	sdv := signExtend(dv, width)
	q := sdd / sdv // Go tronca verso zero, come IDIV
	r := sdd % sdv // resto col segno del dividendo
	lo := -(int64(1) << uint(width-1))
	hi := (int64(1) << uint(width-1)) - 1
	if q < lo || q > hi {
		return 0, 0, false
	}
	return uint16(uint64(q) & mask), uint16(uint64(r) & mask), true
}

func TestMulDifferential(t *testing.T) {
	for _, signed := range []bool{false, true} {
		for a := 0; a <= 0xFF; a++ {
			for b := 0; b <= 0xFF; b++ {
				gotP, gotO := Mul(uint16(a), uint16(b), Width8, signed)
				wantP, wantO := refMul(uint16(a), uint16(b), Width8, signed)
				if gotP != wantP || gotO != wantO {
					t.Fatalf("Mul8 signed=%v a=%#x b=%#x: got(%#x,%v) want(%#x,%v)",
						signed, a, b, gotP, gotO, wantP, wantO)
				}
			}
		}
		// campione a 16 bit
		const step = 433
		for a := 0; a <= 0xFFFF; a += step {
			for b := 0; b <= 0xFFFF; b += step {
				gotP, gotO := Mul(uint16(a), uint16(b), Width16, signed)
				wantP, wantO := refMul(uint16(a), uint16(b), Width16, signed)
				if gotP != wantP || gotO != wantO {
					t.Fatalf("Mul16 signed=%v a=%#x b=%#x: got(%#x,%v) want(%#x,%v)",
						signed, a, b, gotP, gotO, wantP, wantO)
				}
			}
		}
	}
}

func TestDivDifferential(t *testing.T) {
	for _, signed := range []bool{false, true} {
		// 8 bit: dividendo 16 bit campionato, divisore 8 bit completo
		const step8 = 131
		for dd := 0; dd <= 0xFFFF; dd += step8 {
			for dv := 0; dv <= 0xFF; dv++ {
				gotQ, gotR, gotOK := Div(uint32(dd), uint16(dv), Width8, signed)
				wantQ, wantR, wantOK := refDiv(uint32(dd), uint16(dv), Width8, signed)
				if gotOK != wantOK || (gotOK && (gotQ != wantQ || gotR != wantR)) {
					t.Fatalf("Div8 signed=%v dd=%#x dv=%#x: got(%#x,%#x,%v) want(%#x,%#x,%v)",
						signed, dd, dv, gotQ, gotR, gotOK, wantQ, wantR, wantOK)
				}
			}
		}
		// 16 bit: dividendo 32 bit campionato, divisore 16 bit campionato
		const stepDD = 2796203 // primo
		const stepDV = 521
		for dd := uint64(0); dd <= 0xFFFFFFFF; dd += stepDD {
			for dv := 0; dv <= 0xFFFF; dv += stepDV {
				gotQ, gotR, gotOK := Div(uint32(dd), uint16(dv), Width16, signed)
				wantQ, wantR, wantOK := refDiv(uint32(dd), uint16(dv), Width16, signed)
				if gotOK != wantOK || (gotOK && (gotQ != wantQ || gotR != wantR)) {
					t.Fatalf("Div16 signed=%v dd=%#x dv=%#x: got(%#x,%#x,%v) want(%#x,%#x,%v)",
						signed, dd, dv, gotQ, gotR, gotOK, wantQ, wantR, wantOK)
				}
			}
		}
	}
}

// --- Shift/rotate (composti dallo shifter a gate) vs oracolo Go ---

func refShift(op byte, value uint16, count byte, width int, carryIn bool) (uint16, ShiftFlags, bool) {
	mask := uint32(1)<<uint(width) - 1
	v := uint32(value) & mask
	cf := carryIn
	o := op & 0x07
	for i := byte(0); i < count; i++ {
		switch o {
		case ShiftROL:
			top := v >> uint(width-1) & 1
			v = (v<<1 | top) & mask
			cf = top == 1
		case ShiftROR:
			bot := v & 1
			v = v>>1 | bot<<uint(width-1)
			cf = bot == 1
		case ShiftRCL:
			top := v >> uint(width-1) & 1
			ci := b2i(cf)
			v = (v<<1 | ci) & mask
			cf = top == 1
		case ShiftRCR:
			bot := v & 1
			ci := b2i(cf)
			v = v>>1 | ci<<uint(width-1)
			cf = bot == 1
		case ShiftSHR:
			cf = v&1 == 1
			v >>= 1
		case ShiftSAR:
			cf = v&1 == 1
			v = v>>1 | (v>>uint(width-1)&1)<<uint(width-1)
		default: // SHL e alias 6
			cf = v>>uint(width-1)&1 == 1
			v = v << 1 & mask
		}
	}
	res := uint16(v)
	f := ShiftFlags{
		Carry:  cf,
		Sign:   res>>uint(width-1)&1 == 1,
		Zero:   uint32(res)&mask == 0,
		Parity: refParityEven(uint32(res)),
	}
	resMSB := f.Sign
	switch o {
	case ShiftSHL, 6, ShiftROL, ShiftRCL:
		f.Overflow = resMSB != cf
	case ShiftSHR:
		f.Overflow = value>>uint(width-1)&1 == 1
	case ShiftSAR:
		f.Overflow = false
	default: // ROR, RCR
		f.Overflow = resMSB != (res>>uint(width-2)&1 == 1)
	}
	return res, f, o <= ShiftRCR
}

func b2i(v bool) uint32 {
	if v {
		return 1
	}
	return 0
}

func TestShiftDifferential(t *testing.T) {
	ops := []byte{ShiftROL, ShiftROR, ShiftRCL, ShiftRCR, ShiftSHL, ShiftSHR, ShiftSAR, 6}
	for _, width := range []int{Width8, Width16} {
		hi := 0xFF
		if width == Width16 {
			hi = 0xFFFF
		}
		stepV := 1
		if width == Width16 {
			stepV = 257
		}
		for _, op := range ops {
			for v := 0; v <= hi; v += stepV {
				for count := byte(1); count <= 20; count++ {
					for _, cin := range []bool{false, true} {
						gr, gf, giro := Shift(op, uint16(v), count, width, cin)
						wr, wf, wiro := refShift(op, uint16(v), count, width, cin)
						if gr != wr || gf.Carry != wf.Carry || gf.Sign != wf.Sign ||
							gf.Zero != wf.Zero || gf.Parity != wf.Parity || giro != wiro {
							t.Fatalf("op=%d w=%d v=%#x n=%d cin=%v: got(%#x,%+v) want(%#x,%+v)",
								op, width, v, count, cin, gr, gf, wr, wf)
						}
						if count == 1 && gf.Overflow != wf.Overflow {
							t.Fatalf("OF op=%d w=%d v=%#x cin=%v: got %v want %v", op, width, v, cin, gf.Overflow, wf.Overflow)
						}
					}
				}
			}
		}
	}
}

func TestIncDecDifferential(t *testing.T) {
	for _, width := range []int{Width8, Width16} {
		mask := 1<<uint(width) - 1
		for v := 0; v <= mask; v += 1 + v/4096 { // denso in basso, rado in alto
			gotI, fI := Increment(uint16(v), width)
			wantI, wfI := refArith(uint32(v), 1, width, false, false)
			wfI.Carry = false
			if uint32(gotI) != wantI || fI != wfI {
				t.Fatalf("INC width=%d v=%#x: got(%#x,%+v) want(%#x,%+v)", width, v, gotI, fI, wantI, wfI)
			}
			gotD, fD := Decrement(uint16(v), width)
			wantD, wfD := refArith(uint32(v), 1, width, true, true)
			wfD.Carry = false
			if uint32(gotD) != wantD || fD != wfD {
				t.Fatalf("DEC width=%d v=%#x: got(%#x,%+v) want(%#x,%+v)", width, v, gotD, fD, wantD, wfD)
			}
		}
	}
}
