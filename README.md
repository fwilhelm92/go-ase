<!--
SPDX-FileCopyrightText: 2020 SAP SE

SPDX-License-Identifier: Apache-2.0
-->

# go-ase

[![PkgGoDev](https://pkg.go.dev/badge/github.com/SAP/go-ase)](https://pkg.go.dev/github.com/SAP/go-ase)
[![Go Report Card](https://goreportcard.com/badge/github.com/SAP/go-ase)](https://goreportcard.com/report/github.com/SAP/go-ase)
[![REUSE
status](https://api.reuse.software/badge/github.com/SAP/go-ase)](https://api.reuse.software/info/github.com/SAP/go-ase)
![example workflow name](https://github.com/fwilhelm92/go-ase/workflows/CI/badge.svg)


## Description

`go-ase` is a driver for the [`database/sql`][pkg-database-sql] package
of [Go (golang)][go] to provide access to SAP ASE instances.
It is delivered as Go module.

SAP ASE is the shorthand for [SAP Adaptive Server Enterprise][sap-ase],
a relational model database server originally known as Sybase SQL
Server.

A cgo implementation can be found [here][cgo-ase].

## Requirements

The go driver has no special requirements other than Go standard
library and the third part modules listed in `go.mod`, e.g.
`github.com/SAP/go-dblib`.

## Download

The packages in this repo can be `go get` and imported as usual, e.g.:

```sh
go get github.com/SAP/go-ase
```

For specifics on how to use `database/sql` please see the
[documentation][pkg-database-sql].

## Usage

Example code:

```go
package main

import (
    "database/sql"
    _ "github.com/SAP/go-ase"
)

func main() {
    db, err := sql.Open("ase", "ase://user:pass@host:port/")
    if err != nil {
        log.Printf("Failed to open database: %v", err)
        return
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
        log.Printf("Failed to ping database: %v", err)
        return
    }
}
```

### Compilation

```sh
go build -o goase ./cmd/goase/
```

### Execution

```sh
./goase
```

### Examples

More examples can be found in the folder `examples`.

### Integration tests

Integration tests are available and can be run using `go test --tags=integration` and
`go test ./examples/... --tags=integration`.

These require the following environment variables to be set:

- `ASE_HOST`
- `ASE_PORT`
- `ASE_USER`
- `ASE_PASS`

The integration tests will create new databases for each connection type to run tests
against. After the tests are finished the created databases will be removed.

## Configuration

The configuration is handled through either a data source name (DSN) in
one of two forms or through a configuration struct passed to a connector.

All of these support additional properties which can tweak the
connection or the drivers themselves.

### Data Source Names

#### URI DSN

The URI DSN is a common URI like `ase://user:pass@host:port/?prop1=val1&prop2=val2`.

DSNs in this form are parsed using `url.Parse`.

#### Simple DSN

The simple DSN is a key/value string: `username=user password=pass host=hostname port=4901`

Values with spaces must be quoted using single or double quotes.

Each member of `dblib.dsn.DsnInfo` can be set using any of their
possible json tags. E.g. `.Host` will receive the values from the keys
`host` and `hostname`.

Additional properties are set as key/value pairs as well: `...
prop1=val1 prop2=val2`. If the parser doesn't recognize a string as
a json tag it assumes that the key/value pair is a property and its
value.

Similar to the URI DSN those property/value pairs are purely additive.
Any property that only recognizes a single argument (e.g. a boolean)
will only honour the last given value for a property.

#### Connector

As an alternative to the string DSNs `ase.NewConnector` accept a
`dsn.DsnInfo` directly and return a `driver.Connector`, which can 
be passed to `sql.OpenDB`:

```go
package main

import (
    "database/sql"

    "github.com/SAP/go-dblib/dsn"
    "github.com/SAP/go-ase"
)

func main() {
    d := dsn.NewDsnInfo()
    d.Host = "hostname"
    d.Port = "4901"
    d.Username = "user"
    d.Password = "pass"

    connector, err := ase.NewConnector(*d)
    if err != nil {
        log.Printf("Failed to create connector: %v", err)
        return
    }

    db, err := sql.OpenDB(connector)
    if err != nil {
        log.Printf("Failed to open database: %v", err)
        return
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
        log.Printf("Failed to ping ASE: %v", err)
    }
}
```

Additional properties can be set by calling `d.ConnectProps.Add("prop1",
"value1")` or `d.ConnectProps.Set("prop2", "value2")`.

### Properties

##### appname

Recognized values: string

Sets the application name to the value. This can be used in ASE to
determine which application opened a connection.

Defaults to `database/sql driver github.com/SAP/go-ase/purego`.

##### read-only

Recognized values: string

If the value is recognized by `strconv.ParseBool` to represent `true`
the connection will be created as read only.

##### network

Recognized values: string

The network must be a network type recognized by `net.Dial` - at the
time of writing this is either `udp` or `tcp`.

This should only be required to be set if the database is only reachable
through a UDP proxy.

Defaults to `tcp`.

##### channel-package-queue-size

Recognized values: integer

Defines how many packages a TDS channel can buffer at most. When working
with very large datasets where heavy computation only occurs every
hundred packages or so it may be feasible to improve performance by
increasing the queue size.

Defaults to 100.

##### client-hostname

Recognized values: string

The client-hostname to report to the TDS server. Due to protocol
limitations this will be cut off after 30 characters.

Defaults to the hostname of the machine, acquired using `os.Hostname`.

##### packet-read-timeout

Recognized values: integer

The timeout in seconds when a packet is read. The timeout is reset every
time a packet successfully reads data from the connection.

That means the timeout only triggers if no data was read for longer than
`packet-read-timeout` seconds.

Default to 50.

##### tls

Recognized values: bool

Activates TLS for the connection. Any other TLS option is ignored unless tls is set to true.

Defaults to true if the target port is 443, false otherwise.

##### tls-hostname

Recognized values: string

Allows to pass SAN for TLS validation.

For compatibility with the cgo implementation you may also use `ssl`
instead of `tls-hostname` and pass `CN=<SAN>` instead of `<SAN>`.

Defaults to empty string.

Please note that as of go1.15 the CommonName in x509 certificates is no
longer recognized as the hostname if no SANs are present in the
certificate.
If the certificate for your TDS server only utilizes the CN you can
reenable this behaviour by setting `GODEBUG` to `x509ignoreCN=0` in your
environment.

For details see https://golang.google.cn/doc/go1.15#commonname

##### tls-skip-validation

Recognized values: string

If the value is recognized by `strconv.ParseBool` to represent `true`
the TLS certificate of the TDS server will not be validated.

Defaults to empty string / false.

##### tls-ca

Recognized values: string

Path to a CA file, which may contain multiple CAs, to validate the TDS
servers certificate against.
If empty the servers trust store is used.

Defaults to empty string.

## Limitations

### Beta

The go implementation is currently in beta and under active development.
As such most features of the TDS protocol and ASE are not supported.

### Prepared statements

Regarding the limitations of prepared statements/dynamic SQL please see
[the Client-Library documentation](https://help.sap.com/viewer/71b47f4a8269411da6d15ed25f5d39b3/LATEST/en-US/bfc531e46db61014bf8f040071e613d7.html).

The Client-Library documentation applies to the go implementation as
these restrictions are imposed by the implementation of dynamic SQL
on the server side.

### Unsupported ASE data types

Currently the following data types are not supported:

- Timestamp
- Univarchar

## Known Issues

The list of known issues is available [here][issues].

## How to obtain support

Feel free to open issues for feature requests, bugs or general feedback [here][issues].

## Contributing

Any help to improve this package is highly appreciated.

For details on how to contribute please see the
[contributing](CONTRIBUTING.md) file.

## License

Copyright (c) 2019-2020 SAP SE or an SAP affiliate company. All rights reserved.
This file is licensed under the Apache License 2.0 except as noted otherwise in the [LICENSE file](LICENSES).

[cgo-ase]: https://github.com/SAP/cgo-ase
[go]: https://golang.org/
[issues]: https://github.com/SAP/go-ase/issues
[pkg-database-sql]: https://golang.org/pkg/database/sql
[sap-ase]: https://www.sap.com/products/sybase-ase.html
