# embed

Hacked together prototype to rewrite `string` or `[]byte` variables and
constants in a package, using the content of a file.

Meant to be used with `go generate`, comming in go 1.4.

## Usage

```bash
$ cd testdata/
$ embed file -var createDbSQL --source create_query.sql
```

or

```go
package testdata

//go:generate embed file -var createDbSQL --source create_query.sql
var createDbSQL string
```

Flags:

* `-dir`: look for Go files in `testdata`, defaults to current directory
* `-var`: set the variable `createDbSQL`
* `--source` or `stdin`: source of content for `createDbSQL`
* `--keep`: creates a new file instead of setting the variable directly in the file

## Installation

### linux

```bash
wget -qO- https://github.com/aybabtme/embed/releases/download/{{.version}}/embed_linux.tar.gz | tar xvz
```

### darwin

```bash
wget -qO- https://github.com/aybabtme/embed/releases/download/{{.version}}/embed_darwin.tar.gz | tar xvz
```

## Example

### With `go generate`

If you have a Go file:

```go
package testdata

//go:generate embed file -var createDbSQL --source create_query.sql
var createDbSQL string
```

And invoke:

```
$ cd testdata/
$ go generate
embed: in file "example.go"; value of "createDbSQL" set
```

Now the file contains:

```go
package testdata

//go:generate embed file -var createDbSQL --source create_query.sql
var createDbSQL = "CREATE DATABASE IF NOT EXISTS hello_world;\n"
```

### Manually

With the following files:

* `testdata/create_query.sql`

```SQL
CREATE DATABASE IF NOT EXISTS hello_world;
```

* `testdata/example.go`

```go
package testdata

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

var createDbSQL string

func CreateDB(dsn string) error {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return err
    }
    _, err := db.Exec(createDbSQL)
    return err
}

```

Running this command:

```bash
$ embed file --keep -dir testdata -var createDbSQL < testdata/create_query.sql
```

Gives you:

* `testdata/generated_example.go`

```go
package testdata

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

var createDbSQL = "CREATE DATABASE IF NOT EXISTS hello_world;\n"

func CreateDB(dsn string) error {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return err
    }
    _, err := db.Exec(createDbSQL)
    return err
}
```
