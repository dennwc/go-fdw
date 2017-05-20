# go_fdw/Makefile

MODULE_big = go_fdw
OBJS = go_fdw.o

SHLIB_LINK = go_fdw.a

EXTENSION = go_fdw
DATA = go_fdw--1.0.sql

REGRESS = go_fdw

EXTRA_CLEAN = go_fdw.a go_fdw.h

PG_CONFIG = pg_config
PGXS := $(shell $(PG_CONFIG) --pgxs)
include $(PGXS)

go: go_fdw.go
	go build -buildmode=c-archive go_fdw.go