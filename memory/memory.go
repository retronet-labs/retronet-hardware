// Package memory fornisce un modello di memoria a byte per la mini-CPU.
//
// A differenza degli altri componenti di RetroNet Hardware, la memoria è
// modellata in modo *comportamentale* (un array di byte), non costruita dai
// gate: una memoria reale a livello di celle sarebbe enorme e poco istruttiva.
// È la stessa scelta che si fa nei corsi di architettura, dove la memoria è una
// "scatola nera" con lettura/scrittura indirizzate.
package memory

// RAM è una memoria a byte indirizzata a partire da 0.
type RAM struct {
	cells []byte
}

// New crea una memoria di size byte, azzerata. Per l'indirizzamento a 8 bit
// della mini-CPU si usa size = 256.
func New(size int) *RAM {
	return &RAM{cells: make([]byte, size)}
}

// Size restituisce la dimensione in byte.
func (m *RAM) Size() int {
	return len(m.cells)
}

// Read legge il byte all'indirizzo addr.
func (m *RAM) Read(addr byte) byte {
	return m.cells[addr]
}

// Write scrive data all'indirizzo addr.
func (m *RAM) Write(addr, data byte) {
	m.cells[addr] = data
}

// Load copia un programma a partire dall'indirizzo 0 (il punto di avvio della
// mini-CPU). Copia al più Size() byte.
func (m *RAM) Load(program []byte) {
	copy(m.cells, program)
}
