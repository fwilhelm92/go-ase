// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"context"
	"database/sql/driver"
	"fmt"
	"sync"

	"github.com/SAP/go-dblib/dsn"
	"github.com/SAP/go-dblib/tds"
)

// Interface satisfaction checks
var (
	_ driver.Conn               = (*Conn)(nil)
	_ driver.ConnPrepareContext = (*Conn)(nil)
	_ driver.ExecerContext      = (*Conn)(nil)
	_ driver.QueryerContext     = (*Conn)(nil)
	_ driver.Pinger             = (*Conn)(nil)
)

// Conn implements the driver.Conn interface.
type Conn struct {
	Conn    *tds.Conn
	Channel *tds.Channel
	DSN     *dsn.Info

	// TODO I don't particularly like locking statements like this
	stmts map[int]*Stmt
	// TODO: iirc conns aren't used in multiple threads at the same time
	stmtLock *sync.RWMutex
}

// NewConn returns a connection with the passed configuration.
func NewConn(ctx context.Context, dsn *dsn.Info) (*Conn, error) {
	return NewConnWithHooks(ctx, dsn, nil, nil)
}

// NewConnWithHooks returns a connection with the passed configuration.
func NewConnWithHooks(ctx context.Context, dsn *dsn.Info, envChangeHooks []tds.EnvChangeHook, eedHooks []tds.EEDHook) (*Conn, error) {
	conn := &Conn{
		stmts:    map[int]*Stmt{},
		stmtLock: &sync.RWMutex{},
	}

	// Cannot pass the passed context along here as tds.NewConn creates
	// a child context from the passed context.
	// Otherwise the context isn't being used, so using
	// context.Background is fine.
	var err error
	conn.Conn, err = tds.NewConn(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("go-ase: error opening connection to TDS server: %w", err)
	}

	conn.Channel, err = conn.Conn.NewChannel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("go-ase: error opening logical channel: %w", err)
	}

	if drv.envChangeHooks != nil {
		if err := conn.Channel.RegisterEnvChangeHooks(drv.envChangeHooks...); err != nil {
			return nil, fmt.Errorf("go-ase: error registering driver EnvChangeHooks: %w", err)
		}
	}

	if envChangeHooks != nil {
		if err := conn.Channel.RegisterEnvChangeHooks(envChangeHooks...); err != nil {
			return nil, fmt.Errorf("go-ase: error registering argument EnvChangeHooks: %w", err)
		}
	}

	if drv.eedHooks != nil {
		if err := conn.Channel.RegisterEEDHooks(drv.eedHooks...); err != nil {
			return nil, fmt.Errorf("go-ase: error registering driver EEDHooks: %w", err)
		}
	}

	if eedHooks != nil {
		if err := conn.Channel.RegisterEEDHooks(eedHooks...); err != nil {
			return nil, fmt.Errorf("go-ase: error registering argument EEDHooks: %w", err)
		}
	}

	loginConfig, err := tds.NewLoginConfig(dsn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("go-ase: error creating login config: %w", err)
	}

	loginConfig.AppName = dsn.PropDefault("appname", "github.com/SAP/go-ase/purego")

	if err := conn.Channel.Login(ctx, loginConfig); err != nil {
		conn.Close()
		return nil, fmt.Errorf("go-ase: error logging in: %w", err)
	}

	// TODO can this be passed another way?
	if dsn.Database != "" {
		if _, err = conn.ExecContext(ctx, "use "+dsn.Database, nil); err != nil {
			return nil, fmt.Errorf("go-ase: error switching to database %s: %w", dsn.Database, err)
		}
	}

	return conn, nil
}

// Close implements the driver.Conn interface.
func (c *Conn) Close() error {
	if err := c.Conn.Close(); err != nil {
		return fmt.Errorf("go-ase: error closing TDS connection: %w", err)
	}

	return nil
}

// ExecContext implements the driver.ExecerContext.
func (c *Conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	rows, result, err := c.GenericExec(ctx, query, args)

	if rows != nil {
		rows.Close()
	}

	return result, err
}

// QueryContext implements the driver.QueryerContext.
func (c *Conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	rows, _, err := c.GenericExec(ctx, query, args)
	return rows, err
}

// Ping implements the driver.Pinger interface.
func (c Conn) Ping(ctx context.Context) error {
	// TODO implement ErrBadConn check
	rows, _, err := c.language(ctx, "select 'ping'")
	if err != nil {
		return fmt.Errorf("go-ase: error pinging database: %w", err)
	}

	if err := rows.Close(); err != nil {
		return fmt.Errorf("go-ase: error closing rows from ping: %w", err)
	}

	return nil
}
