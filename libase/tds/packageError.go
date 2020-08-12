package tds

import "fmt"

type ErrorPackage struct {
	ErrorNumber int32
	State       uint8
	Class       uint8
	ErrorMsg    string
	ServerName  string
	ProcName    string
	LineNr      uint16
}

func (pkg *ErrorPackage) ReadFrom(ch BytesChannel) error {
	expectLength, err := ch.Uint16()
	if err != nil {
		return ErrNotEnoughBytes
	}

	pkg.ErrorNumber, err = ch.Int32()
	if err != nil {
		return ErrNotEnoughBytes
	}
	n := 4

	msgLength, err := ch.Uint16()
	if err != nil {
		return ErrNotEnoughBytes
	}
	n += 2

	pkg.ErrorMsg, err = ch.String(int(msgLength))
	if err != nil {
		return ErrNotEnoughBytes
	}
	n += int(msgLength)

	serverLength, err := ch.Uint8()
	if err != nil {
		return ErrNotEnoughBytes
	}
	n++

	pkg.ServerName, err = ch.String(int(serverLength))
	if err != nil {
		return ErrNotEnoughBytes
	}
	n += int(serverLength)

	procLength, err := ch.Uint8()
	if err != nil {
		return ErrNotEnoughBytes
	}
	n++

	pkg.ProcName, err = ch.String(int(procLength))
	if err != nil {
		return ErrNotEnoughBytes
	}
	n += int(procLength)

	pkg.LineNr, err = ch.Uint16()
	if err != nil {
		return ErrNotEnoughBytes
	}
	n += 2

	if n != int(expectLength) {
		return fmt.Errorf("expected to read %d bytes, read %d bytes instead",
			expectLength, n)
	}

	return nil
}

func (pkg ErrorPackage) WriteTo(ch BytesChannel) error {
	err := ch.WriteByte(byte(TDS_ERROR))
	if err != nil {
		return fmt.Errorf("failed to write TDS Token %s: %w", TDS_ERROR, err)
	}

	// 4 errornumber
	// 1 state
	// 1 class
	// 2 len(errormsg)
	// x errormsg
	// 1 len(servername)
	// x servername
	// 1 len(procname)
	// x procname
	// 2 linenr
	expectLength := 12 + len(pkg.ErrorMsg) + len(pkg.ServerName) + len(pkg.ProcName)

	err = ch.WriteUint16(uint16(expectLength))
	if err != nil {
		return fmt.Errorf("failed to write length: %w", err)
	}

	err = ch.WriteInt32(pkg.ErrorNumber)
	if err != nil {
		return fmt.Errorf("failed to write error number: %w", err)
	}
	n := 4

	err = ch.WriteUint8(pkg.State)
	if err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}
	n++

	err = ch.WriteUint8(pkg.Class)
	if err != nil {
		return fmt.Errorf("failed to write class: %w", err)
	}
	n++

	err = ch.WriteUint16(uint16(len(pkg.ErrorMsg)))
	if err != nil {
		return fmt.Errorf("failed to write error message length: %w", err)
	}
	n += 2

	err = ch.WriteString(pkg.ErrorMsg)
	if err != nil {
		return fmt.Errorf("failed to write error message: %w", err)
	}
	n += len(pkg.ErrorMsg)

	err = ch.WriteUint8(uint8(len(pkg.ServerName)))
	if err != nil {
		return fmt.Errorf("failed to write servername length: %w", err)
	}
	n++

	err = ch.WriteString(pkg.ServerName)
	if err != nil {
		return fmt.Errorf("failed to write servername: %w", err)
	}
	n += len(pkg.ServerName)

	err = ch.WriteUint8(uint8(len(pkg.ProcName)))
	if err != nil {
		return fmt.Errorf("failed to write procname length: %w", err)
	}
	n++

	err = ch.WriteString(pkg.ProcName)
	if err != nil {
		return fmt.Errorf("failed to write procname: %w", err)
	}
	n += len(pkg.ProcName)

	err = ch.WriteUint16(pkg.LineNr)
	if err != nil {
		return fmt.Errorf("failed to write linenr: %w", err)
	}
	n += 2

	if n != expectLength {
		return fmt.Errorf("expected to write %d bytes, wrote %d bytes instead",
			expectLength, n)
	}

	return nil
}

func (pkg ErrorPackage) String() string {
	return fmt.Sprintf("%T(%d: %s)", pkg, pkg.ErrorNumber, pkg.ErrorMsg)
}
