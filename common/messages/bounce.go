// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import ()
import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"time"
	"strings"
)

type Bounce struct {
	MessageBase
	Name      string
	Timestamp interfaces.Timestamp
	Stamps    []interfaces.Timestamp
	size      int
}

var _ interfaces.IMsg = (*Bounce)(nil)

func (m *Bounce) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

// We have to return the haswh of the underlying message.
func (m *Bounce) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *Bounce) SizeOf() int {
	m.GetMsgHash()
	return m.size
}

func (m *Bounce) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalForSignature()

		m.size = len(data)

		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *Bounce) Type() byte {
	return constants.BOUNCE_MSG
}

func (m *Bounce) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *Bounce) VerifySignature() (bool, error) {
	return true, nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *Bounce) Validate(state interfaces.IState) int {
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *Bounce) ComputeVMIndex(state interfaces.IState) {

}

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *Bounce) LeaderExecute(state interfaces.IState) {
}

func (m *Bounce) FollowerExecute(state interfaces.IState) {
}

// Acknowledgements do not go into the process list.
func (e *Bounce) Process(dbheight uint32, state interfaces.IState) bool {
	return true
}

func (e *Bounce) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *Bounce) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *Bounce) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (m *Bounce) Sign(key interfaces.Signer) error {
	return nil
}

func (m *Bounce) GetSignature() interfaces.IFullSignature {
	return nil
}

func (m *Bounce) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data
	if newData[0] != m.Type() {
		return nil, errors.New("Invalid Message type")
	}
	newData = newData[1:]

	m.Name = string(newData[:32])
	newData = newData[32:]

	m.Timestamp = new(primitives.Timestamp)
	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	numTS, newData := binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	for i := uint32(0); i < numTS; i++ {
		ts := new(primitives.Timestamp)
		newData, err = ts.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
		m.Stamps = append(m.Stamps, ts)
	}
	return
}

func (m *Bounce) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *Bounce) MarshalForSignature() ([]byte, error) {
	var buf primitives.Buffer

	binary.Write(&buf, binary.BigEndian, m.Type())

	var buff [32]byte

	copy(buff[:32], []byte(fmt.Sprintf("%32s", m.Name)))
	buf.Write(buff[:])

	t := m.GetTimestamp()
	data, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, int32(len(m.Stamps)))

	for _, ts := range m.Stamps {
		data, err := ts.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(data)
	}

	return buf.DeepCopyBytes(), nil
}

func (m *Bounce) MarshalBinary() (data []byte, err error) {
	return m.MarshalForSignature()
}

func (m *Bounce) String() string {
	str := fmt.Sprintf("\nbbbb Origin: %s %s Bounce Start:  %30s Hops: %5d Size: %5d ", time.Now().String(), strings.TrimSpace(m.Name), m.Timestamp.String(),len(m.Stamps),m.SizeOf())
	last := m.Timestamp.GetTimeMilli()
	elapse := int64(0)
	sum := elapse
	for _, ts := range m.Stamps {
		elapse = ts.GetTimeMilli() - last
		sum += elapse
		last = ts.GetTimeMilli()
//		str = fmt.Sprintf("%sbbbb %30s %4d.%03d seconds\n", str, ts.String(), elapse/1000, elapse%1000)
	}
	if len(m.Stamps) > 0 {
		avg := sum/int64(len(m.Stamps))
		str = str + fmt.Sprintf("Last Hop Took %d.%03d Average Hop: %d.%03d\n",elapse/1000,elapse%1000,avg/1000,avg%1000)
	}
	return str
}

func (a *Bounce) IsSameAs(b *Bounce) bool {

	return true
}
