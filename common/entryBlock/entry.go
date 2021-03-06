// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryBlock

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// An Entry is the element which carries user data
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#entry
type Entry struct {
	Version uint8
	ChainID interfaces.IHash
	ExtIDs  [][]byte
	Content []byte
}

var _ interfaces.IEBEntry = (*Entry)(nil)
var _ interfaces.DatabaseBatchable = (*Entry)(nil)
var _ interfaces.BinaryMarshallable = (*Entry)(nil)

// Returns the size of the entry subject to payment in K.  So anything up
// to 1K returns 1, everything up to and including 2K returns 2, etc.
// An error returns 100 (an invalid size)
func (c *Entry) KSize() int {
	data, err := c.MarshalBinary()
	if err != nil {
		return 100
	}
	return (len(data) - 35 + 1023) / 1024
}

func (c *Entry) New() interfaces.BinaryMarshallableAndCopyable {
	return NewEntry()
}

func (c *Entry) GetDatabaseHeight() uint32 {
	return 0
}

func (e *Entry) GetWeld() []byte {
	return primitives.DoubleSha(append(e.GetHash().Bytes(), e.GetChainID().Bytes()...))
}

func (e *Entry) GetWeldHash() interfaces.IHash {
	hash := primitives.NewZeroHash()
	hash.SetBytes(e.GetWeld())
	return hash
}

func (c *Entry) GetChainID() interfaces.IHash {
	return c.ChainID
}

func (c *Entry) DatabasePrimaryIndex() interfaces.IHash {
	return c.GetHash()
}

func (c *Entry) DatabaseSecondaryIndex() interfaces.IHash {
	return nil
}

// NewChainID generates a ChainID from an entry. ChainID = primitives.Sha(Sha(ExtIDs[0]) +
// Sha(ExtIDs[1] + ... + Sha(ExtIDs[n]))
func NewChainID(e interfaces.IEBEntry) interfaces.IHash {
	id := new(primitives.Hash)
	sum := sha256.New()
	for _, v := range e.ExternalIDs() {
		x := sha256.Sum256(v)
		sum.Write(x[:])
	}
	id.SetBytes(sum.Sum(nil))

	return id
}

func (e *Entry) GetContent() []byte {
	return e.Content
}

func (e *Entry) GetChainIDHash() interfaces.IHash {
	return e.ChainID
}

func (e *Entry) ExternalIDs() [][]byte {
	return e.ExtIDs
}

func (e *Entry) IsValid() bool {

	//double check the version
	if e.Version != 0 {
		return false
	}

	return true
}

func (e *Entry) GetHash() interfaces.IHash {
	h := primitives.NewZeroHash()
	entry, err := e.MarshalBinary()
	if err != nil {
		return h
	}

	h1 := sha512.Sum512(entry)
	h2 := sha256.Sum256(append(h1[:], entry[:]...))
	h.SetBytes(h2[:])
	return h
}

func (e *Entry) MarshalBinary() ([]byte, error) {
	buf := new(primitives.Buffer)

	// 1 byte Version
	if err := binary.Write(buf, binary.BigEndian, e.Version); err != nil {
		return nil, err
	}

	// 32 byte ChainID
	buf.Write(e.ChainID.Bytes())

	// ExtIDs
	if ext, err := e.MarshalExtIDsBinary(); err != nil {
		return nil, err
	} else {
		// 2 byte size of ExtIDs
		if err := binary.Write(buf, binary.BigEndian, int16(len(ext))); err != nil {
			return nil, err
		}

		// binary ExtIDs
		buf.Write(ext)
	}

	// Content
	buf.Write(e.Content)

	return buf.DeepCopyBytes(), nil
}

// MarshalExtIDsBinary marshals the ExtIDs into a []byte containing a series of
// 2 byte size of each ExtID followed by the ExtID.
func (e *Entry) MarshalExtIDsBinary() ([]byte, error) {
	buf := new(primitives.Buffer)

	for _, x := range e.ExtIDs {
		// 2 byte size of the ExtID
		if err := binary.Write(buf, binary.BigEndian, uint16(len(x))); err != nil {
			return nil, err
		}

		// ExtID bytes
		buf.Write(x)
	}

	return buf.DeepCopyBytes(), nil
}

func UnmarshalEntry(data []byte) (interfaces.IEBEntry, error) {
	entry := NewEntry()
	err := entry.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (e *Entry) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	buf := primitives.NewBuffer(data)
	hash := make([]byte, 32)

	// 1 byte Version
	b, err := buf.ReadByte()
	if err != nil {
		return nil, err
	} else {
		e.Version = b
	}

	// 32 byte ChainID
	e.ChainID = primitives.NewZeroHash()
	if _, err = buf.Read(hash); err != nil {
		return nil, err
	} else if err = e.ChainID.SetBytes(hash); err != nil {
		return nil, err
	}

	// 2 byte size of ExtIDs
	var extSize uint16
	if err = binary.Read(buf, binary.BigEndian, &extSize); err != nil {
		return nil, err
	}

	// ExtIDs
	for i := int16(extSize); i > 0; {
		var xsize int16
		binary.Read(buf, binary.BigEndian, &xsize)
		i -= 2
		if i < 0 {
			err = fmt.Errorf("Error parsing external IDs")
			return nil, err
		}
		x := make([]byte, xsize)
		var n int
		if n, err = buf.Read(x); err != nil {
			return nil, err
		} else {
			if c := cap(x); n != c {
				err = fmt.Errorf("Could not read ExtID: Read %d bytes of %d\n", n, c)
				return nil, err
			}
			e.ExtIDs = append(e.ExtIDs, x)
			i -= int16(n)
			if i < 0 {
				err = fmt.Errorf("Error parsing external IDs")
				return nil, err
			}
		}
	}

	// Content
	e.Content = buf.DeepCopyBytes()

	return
}

func (e *Entry) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *Entry) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *Entry) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *Entry) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *Entry) String() string {
	str, _ := e.JSONString()
	return str
}

/***************************************************************
 * Helper Functions
 ***************************************************************/

func NewEntry() *Entry {
	e := new(Entry)
	e.ChainID = primitives.NewZeroHash()
	e.ExtIDs = make([][]byte, 0)
	e.Content = make([]byte, 0)
	return e
}
