package i8080

import "testing"

func TestALUAddSetsCarryAndAuxiliaryCarry(t *testing.T) {
	result, flags := ALU(GroupADD, 0x8F, 0x81, false)
	if result != 0x10 || !flags.Carry || !flags.AuxiliaryCarry {
		t.Fatalf("ADD result=0x%02X flags=%+v", result, flags)
	}
}

func TestALUSubUsesBorrowConvention(t *testing.T) {
	result, flags := ALU(GroupSUB, 0x00, 0x01, false)
	if result != 0xFF || !flags.Carry || !flags.AuxiliaryCarry || !flags.Sign {
		t.Fatalf("SUB result=0x%02X flags=%+v", result, flags)
	}
}

func TestLogicalFlagsMatch8080Convention(t *testing.T) {
	result, flags := ALU(GroupANA, 0xF0, 0x0F, false)
	if result != 0 || flags.Carry || !flags.AuxiliaryCarry || !flags.Zero {
		t.Fatalf("ANA result=0x%02X flags=%+v", result, flags)
	}

	result, flags = ALU(GroupXRA, 0x55, 0x55, false)
	if result != 0 || flags.Carry || flags.AuxiliaryCarry || !flags.Zero {
		t.Fatalf("XRA result=0x%02X flags=%+v", result, flags)
	}
}

func TestAdd16CascadesCarry(t *testing.T) {
	result, carry := Add16(0xFFFF, 0x0001)
	if result != 0x0000 || !carry {
		t.Fatalf("Add16 result=0x%04X carry=%v", result, carry)
	}
}
