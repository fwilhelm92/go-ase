// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package tds

import (
	"errors"
	"fmt"
	"strconv"
)

//go:generate stringer -type=DynamicOperationType
type DynamicOperationType byte

const (
	TDS_DYN_INVALID    DynamicOperationType = 0x00
	TDS_DYN_PREPARE    DynamicOperationType = 0x01
	TDS_DYN_EXEC       DynamicOperationType = 0x02
	TDS_DYN_DEALLOC    DynamicOperationType = 0x04
	TDS_DYN_EXEC_IMMED DynamicOperationType = 0x08
	TDS_DYN_PROCNAME   DynamicOperationType = 0x10
	TDS_DYN_ACK        DynamicOperationType = 0x20
	TDS_DYN_DESCIN     DynamicOperationType = 0x40
	TDS_DYN_DESCOUT    DynamicOperationType = 0x80
)

//go:generate stringer -type=DynamicStatusType
type DynamicStatusType byte

const (
	TDS_DYNAMIC_UNUSED            DynamicStatusType = 0x00
	TDS_DYNAMIC_HASARGS           DynamicStatusType = 0x01
	TDS_DYNAMIC_SUPPRESS_FMT      DynamicStatusType = 0x2
	TDS_DYNAMIC_BATCH_PARAMS      DynamicStatusType = 0x4
	TDS_DYNAMIC_SUPPRESS_PARAMFMT DynamicStatusType = 0x08
)

type DynamicPackage struct {
	Type   DynamicOperationType
	Status DynamicStatusType
	ID     int
	Stmt   string

	wide bool
}

func (pkg *DynamicPackage) ReadFrom(ch BytesChannel) error {
	var totalLength int
	var err error
	if pkg.wide {
		var length uint32
		length, err = ch.Uint32()
		totalLength = int(length)
	} else {
		var length uint16
		length, err = ch.Uint16()
		totalLength = int(length)
	}
	if err != nil {
		return ErrNotEnoughBytes
	}

	dynamicType, err := ch.Byte()
	if err != nil {
		return ErrNotEnoughBytes
	}
	pkg.Type = DynamicOperationType(dynamicType)
	n := 1

	dynamicStatus, err := ch.Byte()
	if err != nil {
		return ErrNotEnoughBytes
	}
	pkg.Status = DynamicStatusType(dynamicStatus)
	n++

	idLen, err := ch.Uint8()
	if err != nil {
		return ErrNotEnoughBytes
	}
	n++

	id, err := ch.String(int(idLen))
	if err != nil {
		return ErrNotEnoughBytes
	}
	n += int(idLen)

	pkg.ID, err = strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("error parsing dynamic statement ID '%s': %w", id, err)
	}

	if pkg.Type&TDS_DYN_PREPARE == TDS_DYN_PREPARE || pkg.Type&TDS_DYN_EXEC_IMMED == TDS_DYN_EXEC_IMMED {
		var stmtLen int
		if pkg.wide {
			var length uint32
			length, err = ch.Uint32()
			stmtLen = int(length)
			n += 4
		} else {
			var length uint16
			length, err = ch.Uint16()
			stmtLen = int(length)
			n += 2
		}
		if err != nil {
			return ErrNotEnoughBytes
		}

		pkg.Stmt, err = ch.String(int(stmtLen))
		if err != nil {
			return ErrNotEnoughBytes
		}
		n += int(stmtLen)
	}

	if n != totalLength {
		return fmt.Errorf("expected to read %d bytes, read %d bytes instead", totalLength, n)
	}

	return nil
}

func (pkg *DynamicPackage) WriteTo(ch BytesChannel) error {
	if pkg.Type == TDS_DYN_INVALID {
		return errors.New("dynamic type is invalid")
	}

	token := TDS_DYNAMIC
	if pkg.wide {
		token = TDS_DYNAMIC2
	}

	if err := ch.WriteByte(byte(token)); err != nil {
		return err
	}

	strId := strconv.FormatInt(int64(pkg.ID), 10)

	// 1  dynamicType
	// 1  dynamicStatus
	// 1  id length
	// x  id
	// 2  stmt length if !pkg.wide
	// 4 stmt length if pkg.wide
	// x  stmt
	totalLength := 5 + len(strId) + len(pkg.Stmt)
	if pkg.wide {
		// add two more bytes for TDS_DYNAMIC2
		totalLength += 2
	}

	if err := ch.WriteUint16(uint16(totalLength)); err != nil {
		return err
	}

	if err := ch.WriteByte(byte(pkg.Type)); err != nil {
		return err
	}
	n := 1

	if err := ch.WriteByte(byte(pkg.Status)); err != nil {
		return err
	}
	n++

	if err := ch.WriteUint8(uint8(len(strId))); err != nil {
		return err
	}
	n++

	if err := ch.WriteString(strId); err != nil {
		return err
	}
	n += len(strId)

	if pkg.Type&TDS_DYN_PREPARE == TDS_DYN_PREPARE || pkg.Type&TDS_DYN_EXEC_IMMED == TDS_DYN_EXEC_IMMED {
		if pkg.wide {
			if err := ch.WriteUint32(uint32(len(pkg.Stmt))); err != nil {
				return err
			}
			n += 4
		} else {
			if err := ch.WriteUint16(uint16(len(pkg.Stmt))); err != nil {
				return err
			}
			n += 2
		}

		if err := ch.WriteString(pkg.Stmt); err != nil {
			return err
		}
		n += len(pkg.Stmt)
	}

	if n != totalLength {
		return fmt.Errorf("expected to write %d bytes, wrote %d bytes instead", totalLength, n)
	}

	return nil
}

func (pkg DynamicPackage) String() string {
	types := deBitmask(int(pkg.Type), int(TDS_DYN_DESCOUT))
	strTypes := ""
	if len(types) == 0 {
		strTypes = TDS_DYN_PREPARE.String()
	} else {
		for i, typ := range types {
			strTypes += DynamicOperationType(typ).String()
			if i+1 != len(types) {
				strTypes += "|"
			}
		}
	}

	stati := deBitmask(int(pkg.Status), int(TDS_DYNAMIC_SUPPRESS_FMT))
	strStati := ""
	if len(stati) == 0 {
		strStati = TDS_DYNAMIC_UNUSED.String()
	} else {
		for i, status := range stati {
			strStati += DynamicStatusType(status).String()
			if i+1 != len(stati) {
				strStati += "|"
			}
		}
	}

	return fmt.Sprintf("%T(%s, %s - %d: %s)", pkg, strTypes, strStati, pkg.ID, pkg.Stmt)
}
