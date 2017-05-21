# Go FDW for PostgreSQL

An experimental Go project template for building PostgreSQL Foreign Data Wrappers (FDW).

Tested with PostgreSQL v9.6 and Go 1.8.1 on Ubuntu x64.

**Supports:**

* Table scan
* EXPLAIN
* Table options

Contributions are welcome!

## Getting started

Module entry point is defined in `fdw.go` file (see `SetTable`). This file contains a basic working example, so give it a try.
Later you will need to rewrite it to suit your needs.

### Build

The easiest way to build Postgres extension is to run `dennwc/go_fdw` Docker image:

```
docker run --rm -v $GOPATH:/gopath -v $PWD/fdw.go:/build/fdw.go -v $PWD/out:/out dennwc/go_fdw
```

This command will mount your `./fdw.go` and `GOPATH` to the container, build an extension and copy it to `./out` folder.

If you don't use Docker, check `Dockerfile` to see what commands are needed to manually setup your build environment.

### Install

To install an extension just copy it to your Postgres installation. At the end of the build process you'll see following lines:

```
/usr/bin/install -c -m 755  go_fdw.so '/usr/lib/postgresql/9.6/lib/go_fdw.so'
/usr/bin/install -c -m 644 .//go_fdw.control '/usr/share/postgresql/9.6/extension/'
/usr/bin/install -c -m 644 .//go_fdw--1.0.sql  '/usr/share/postgresql/9.6/extension/'
```

You can execute the same commands to install an extension locally.

### Testing

Execute following SQL statements to load an extension:

```sql
-- should match an extension lib name
-- creates a "go-fdw" FDW server record automatically (see go_fdw--1.0.sql)
CREATE EXTENSION go_fdw;

-- create a foreign table for an extension
CREATE FOREIGN TABLE public.gotest (
  id INTEGER NOT NULL,
  name text NOT NULL
)
SERVER "go-fdw"
OPTIONS (foo 'bar');
```

And finally, run the query:

```sql
SELECT * FROM gotest;
```

**Note:** After extension is loaded, you'll need to restart Postgres to test any changes to the code.
Database will not reload shared library file automatically.

## Hacking

Project is under development and lacks few features, so feel free to send a PR or file an issue :)

An official documentation for FDW callbacks can be found [here](https://www.postgresql.org/docs/9.6/static/fdwhandler.html).
And the documentation for Postgres sources is [here](https://doxygen.postgresql.org).

