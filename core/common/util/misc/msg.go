package misc

import (
	"encoding/binary"
	"errors"
)

//	=====================================================================
//	function name: AddLengthForMsg
//	function type: public
//	function receiver: na
//	calculate the length of the message byte array, then append it,
//	return messages length and message into byte array
//	=====================================================================
func AddLengthForMsg(msgs ...[]byte) []byte {
	result := []byte{}
	if len(msgs) == 0 {
		return nil
	}

	for _, msg := range msgs {
		if msg == nil || len(msg) == 0 {
			continue
		}
		length := make([]byte, 4)
		binary.BigEndian.PutUint32(length, uint32(len(msg)))
		tmp := append(length, msg...)
		result = append(result, tmp...)
	}

	return result
}

//	=====================================================================
//	function name: GetWithLengthMsg
//	function type: public
//	function receiver: na
//	get the message byte array include message length, then remove the
//	message length, return the message byte array with no length
//	=====================================================================
func GetWithLengthMsg(meta []byte) ([][]byte, error) {
	result := [][]byte{}
	if len(meta) == 0 {
		return nil, errors.New("no meta data")
	}

	tmp := meta[:]
	for {
		if len(tmp) < 4 {
			return nil, errors.New("data not format")
		}

		length := binary.BigEndian.Uint32(tmp[:4])
		if int(length+4) > len(tmp) {
			return nil, errors.New("data not format")
		}

		data := tmp[4 : length+4]
		result = append(result, data)
		if len(tmp) == int(length+4) {
			break
		}

		tmp = tmp[length+4:]
	}

	return result, nil
}
