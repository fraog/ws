package ws

import (
	"crypto/rand"
	"encoding/binary"
	"io"
)

const (
	opContinue   = 0x00
	opText       = 0x01
	opBinary     = 0x02
	opControlMin = 0x03
	opControlMax = 0x07
	opClose      = 0x08
	opPing       = 0x09
	opPong       = 0x0A
	selectOp     = 0x0F //15
	selectPl     = 0x7f //?
	sigUint16    = 0x7E //126
	sigUint64    = 0x7F //127
	msbOn        = 0x80 //128
	stdResponse  = 0x81 //129
)

type DataFrame struct {
	fin    byte
	op     byte
	pl     byte
	plext  []byte
	masked byte
	mask   []byte
	length int
}

//Construct a dataframe for message.
func NewFrame(message []byte) *DataFrame {
	var f DataFrame
	f = DataFrame{msbOn, 1, 0, nil, 0, nil, 0}
	f.SetDataLength(len(message))
	return &f
}

//Returns the length of the data described by this data frame.
func (f *DataFrame) GetDataLength() int {
	return f.length
}

//Sets the length described by this data frame to l.
func (f *DataFrame) SetDataLength(l int) {
	switch {
	case l <= 125: //UINT8
		f.pl = byte(l)
		f.plext = nil
	case l <= 65535: //UINT16
		f.pl = sigUint16
		f.plext = make([]byte, 2)
		binary.BigEndian.PutUint16(f.plext, uint16(l))
	default: //UINT64
		f.pl = sigUint64
		f.plext = make([]byte, 8)
		binary.BigEndian.PutUint64(f.plext, uint64(l))
	}
	f.length = l
}

/*
Reads a WebSocket data frame from reader into this data frame.
Returns an error if the frame could not be read.
*/
func (f *DataFrame) ReadFrom(r io.Reader) error {

	var e error
	var b []byte
	b = make([]byte, 1)

	//FIN + Opcode
	if _, e = r.Read(b); e != nil {
		return e
	}

	f.fin = (msbOn & b[0])
	f.op = selectOp & b[0] //15

	//Evaluate OpCode
	if f.op > opBinary {
		return nil
	}

	//Get Hash bit and Payload length
	if _, e = r.Read(b); e != nil {
		return e
	}

	f.masked = msbOn & b[0]
	if f.masked != msbOn { //128, MSBON
		//FIXME: Create custom error type for unuhashed
		//TODO: We don't want to end here because what if its a client that is using this? We'll check masked before decoding
		//return e
	}

	f.pl = selectPl & b[0]
	switch f.pl {
	case sigUint16:
		f.plext = make([]byte, 2)
		if _, e = io.ReadFull(r, f.plext); e != nil {
			return e
		}
		f.length = int(binary.BigEndian.Uint16(f.plext))
	case sigUint64:
		f.plext = make([]byte, 8)
		if _, e = io.ReadFull(r, f.plext); e != nil {
			return e
		}
		f.length = int(binary.BigEndian.Uint64(f.plext))
	default:
		f.plext = nil
		f.length = int(0x7F & b[0])
	}

	//Read the mask if it has one
	if f.masked == msbOn {
		if f.mask == nil {
			f.mask = make([]byte, 4)
		}
		if _, e = io.ReadFull(r, f.mask); e != nil {
			return e
		}
	}

	return nil
}

/*
Writes this WebSocket data frame to a writer.
Returns an error if the frame could not be written.
*/
func (f *DataFrame) WriteTo(w io.Writer) error {

	//fmt.Printf("FIN:%b\n", f.fin)
	//fmt.Printf("OP:%b\n", f.op)
	//fmt.Printf("1st:%b\n", (f.fin | f.op))

	var e error
	if _, e = w.Write([]byte{(f.fin | f.op), (f.masked | f.pl)}); e != nil {
		return e
	}
	if f.plext != nil {
		if _, e = w.Write(f.plext); e != nil {
			return e
		}
	}
	if f.masked == msbOn { //0x80, MSB on
		if _, e = w.Write(f.mask); e != nil {
			return e
		}
	}
	return nil
}

/*
Reads an encoded WebSocket payload from reader, decodes it into a byte slice.
Returns an error if the data was not written.
*/
func (f *DataFrame) Decode(r io.Reader, w []byte) error {
	//Read and decode the payload byte by byte
	var e error
	b := make([]byte, 1)
	for i := 0; i < f.length; i++ {
		if _, e = r.Read(b); e != nil {
			return e
		}
		w[i] = b[0] ^ f.mask[i%4]
	}
	return nil
}

/*
Reads an encoded WebSocket payload from reader, decodes it, and writes it to a Writer.
Returns an error if the data was not written.
*/
func (f *DataFrame) DecodeTo(r io.Reader, w io.Writer) error {
	//Read and decode the payload byte by byte
	var e error
	b := make([]byte, 1)
	for i := 0; i < f.length; i++ {
		if _, e = r.Read(b); e != nil {
			return e
		}
		b[0] = b[0] ^ f.mask[i%4]
		if _, e = w.Write(b); e != nil {
			return e
		}
	}
	return nil
}

/*
Generates a random mask to use for encoding a message.
After this method, the data frame's masked bit is set to 1 and mask slice is full.
FIXME: crypto/rand vs rand? This is using crypto/rand.
*/
func (f *DataFrame) GenerateMask() error {
	f.masked = msbOn

	if f.mask == nil {
		f.mask = make([]byte, 4)
	}

	if _, e := rand.Read(f.mask); e != nil {
		return e
	}

	return nil
}

/*
Encodes a message using the current mask, and writes the result to a Writer.
*/
func (f *DataFrame) Encode(m []byte, w io.Writer) error {
	b := make([]byte, 1)
	for i := 0; i < f.length; i++ {
		b[0] = m[i] ^ f.mask[i%4]
		if _, e := w.Write(b); e != nil {
			return e
		}
	}
	return nil
}

/*
Encodes a message from reader using the current mask, the length of which is denoted by the frame, and writes the result to a Writer.
*/
func (f *DataFrame) EncodeTo(r io.Reader, w io.Writer) error {
	b := make([]byte, 1)
	for i := 0; i < f.length; i++ {
		if _, e := r.Read(b); e != nil {
			return e
		}
		b[0] = b[0] ^ f.mask[i%4]
		if _, e := w.Write(b); e != nil {
			return e
		}
	}
	return nil
}
