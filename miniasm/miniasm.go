// Package miniasm è un piccolo assembler per l'ISA della mini-CPU di RetroNet
// Hardware (vedi docs/cpu-isa.md): traduce un sorgente testuale in codice
// macchina, con supporto a etichette, commenti (";") e valori in decimale o
// esadecimale ("0x..").
//
// Funziona in due passate: la prima calcola gli indirizzi e la tabella delle
// etichette, la seconda emette i byte risolvendo gli operandi.
package miniasm

import (
	"fmt"
	"strconv"
	"strings"
)

// form descrive il formato di codifica di un'istruzione.
type form int

const (
	formNone    form = iota // 1 byte: solo opcode
	formReg                 // 1 byte: opcode | Rd<<2
	formRegReg              // 1 byte: opcode | Rd<<2 | Rs
	formRegImm              // 2 byte: [opcode | Rd<<2, immediato]
	formRegAddr             // 2 byte: [opcode | Rd<<2, indirizzo]
	formAddr                // 2 byte: [opcode, indirizzo]
)

type instr struct {
	opcode byte
	form   form
}

// table associa ogni mnemonico al suo opcode base e formato.
var table = map[string]instr{
	"NOP":  {0x00, formNone},
	"LDI":  {0x10, formRegImm},
	"MOV":  {0x20, formRegReg},
	"ADD":  {0x30, formRegReg},
	"SUB":  {0x40, formRegReg},
	"AND":  {0x50, formRegReg},
	"OR":   {0x60, formRegReg},
	"LD":   {0x70, formRegAddr},
	"ST":   {0x80, formRegAddr},
	"JMP":  {0x90, formAddr},
	"JZ":   {0xA0, formAddr},
	"JC":   {0xB0, formAddr},
	"SHL":  {0xC0, formReg},
	"SHR":  {0xD0, formReg},
	"CALL": {0xE1, formAddr},
	"RET":  {0xE0, formNone},
	"HLT":  {0xF0, formNone},
}

func (f form) size() int {
	switch f {
	case formRegImm, formRegAddr, formAddr:
		return 2
	default:
		return 1
	}
}

type stmt struct {
	num      int
	mnem     string
	operands []string
}

// Assemble traduce il sorgente assembly in codice macchina per la mini-CPU.
func Assemble(src string) ([]byte, error) {
	stmts, labels, err := pass1(src)
	if err != nil {
		return nil, err
	}
	return pass2(stmts, labels)
}

// pass1 estrae le istruzioni, calcola gli indirizzi e raccoglie le etichette.
func pass1(src string) ([]stmt, map[string]byte, error) {
	labels := map[string]byte{}
	var stmts []stmt
	addr := 0

	for i, raw := range strings.Split(src, "\n") {
		num := i + 1
		text := raw
		if idx := strings.IndexByte(text, ';'); idx >= 0 {
			text = text[:idx] // rimuove il commento
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		// Etichetta opzionale a inizio riga: "nome:".
		if idx := strings.IndexByte(text, ':'); idx >= 0 {
			label := strings.TrimSpace(text[:idx])
			if !validLabel(label) {
				return nil, nil, fmt.Errorf("riga %d: etichetta non valida %q", num, label)
			}
			if _, dup := labels[label]; dup {
				return nil, nil, fmt.Errorf("riga %d: etichetta duplicata %q", num, label)
			}
			labels[label] = byte(addr)
			text = strings.TrimSpace(text[idx+1:])
			if text == "" {
				continue
			}
		}

		fields := strings.Fields(text)
		mnem := strings.ToUpper(fields[0])
		ins, ok := table[mnem]
		if !ok {
			return nil, nil, fmt.Errorf("riga %d: istruzione sconosciuta %q", num, fields[0])
		}

		rest := strings.TrimSpace(text[len(fields[0]):])
		var operands []string
		if rest != "" {
			for _, p := range strings.Split(rest, ",") {
				operands = append(operands, strings.TrimSpace(p))
			}
		}

		stmts = append(stmts, stmt{num: num, mnem: mnem, operands: operands})
		addr += ins.form.size()
		if addr > 256 {
			return nil, nil, fmt.Errorf("riga %d: il programma supera i 256 byte", num)
		}
	}
	return stmts, labels, nil
}

// pass2 emette i byte risolvendo registri, immediati, indirizzi ed etichette.
func pass2(stmts []stmt, labels map[string]byte) ([]byte, error) {
	var code []byte
	for _, s := range stmts {
		ins := table[s.mnem]
		switch ins.form {
		case formNone:
			if err := want(s, 0); err != nil {
				return nil, err
			}
			code = append(code, ins.opcode)
		case formReg:
			if err := want(s, 1); err != nil {
				return nil, err
			}
			rd, err := parseReg(s, s.operands[0])
			if err != nil {
				return nil, err
			}
			code = append(code, ins.opcode|rd<<2)
		case formRegReg:
			if err := want(s, 2); err != nil {
				return nil, err
			}
			rd, err := parseReg(s, s.operands[0])
			if err != nil {
				return nil, err
			}
			rs, err := parseReg(s, s.operands[1])
			if err != nil {
				return nil, err
			}
			code = append(code, ins.opcode|rd<<2|rs)
		case formRegImm, formRegAddr:
			if err := want(s, 2); err != nil {
				return nil, err
			}
			rd, err := parseReg(s, s.operands[0])
			if err != nil {
				return nil, err
			}
			v, err := parseValue(s, s.operands[1], labels)
			if err != nil {
				return nil, err
			}
			code = append(code, ins.opcode|rd<<2, v)
		case formAddr:
			if err := want(s, 1); err != nil {
				return nil, err
			}
			v, err := parseValue(s, s.operands[0], labels)
			if err != nil {
				return nil, err
			}
			code = append(code, ins.opcode, v)
		}
	}
	return code, nil
}

func want(s stmt, n int) error {
	if len(s.operands) != n {
		return fmt.Errorf("riga %d: %s richiede %d operandi, trovati %d", s.num, s.mnem, n, len(s.operands))
	}
	return nil
}

func parseReg(s stmt, tok string) (byte, error) {
	t := strings.ToUpper(tok)
	if len(t) == 2 && t[0] == 'R' && t[1] >= '0' && t[1] <= '3' {
		return t[1] - '0', nil
	}
	return 0, fmt.Errorf("riga %d: registro non valido %q (atteso R0-R3)", s.num, tok)
}

func parseValue(s stmt, tok string, labels map[string]byte) (byte, error) {
	if v, ok := labels[tok]; ok {
		return v, nil
	}
	n, err := strconv.ParseUint(tok, 0, 8) // base 0: accetta 0x.. e decimale
	if err != nil {
		return 0, fmt.Errorf("riga %d: valore o etichetta non valido %q", s.num, tok)
	}
	return byte(n), nil
}

func validLabel(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		switch {
		case r == '_', r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		case i > 0 && r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return true
}
