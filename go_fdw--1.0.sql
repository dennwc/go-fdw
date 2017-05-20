/* go_fdw/go_fdw--1.0.sql */

-- complain if script is sourced in psql, rather than via CREATE EXTENSION
\echo Use "CREATE EXTENSION go_fdw" to load this extension. \quit

CREATE FUNCTION go_fdw_handler()
RETURNS fdw_handler
AS 'MODULE_PATHNAME'
LANGUAGE C STRICT;

CREATE FUNCTION go_fdw_validator(text[], oid)
RETURNS void
AS 'MODULE_PATHNAME'
LANGUAGE C STRICT;

CREATE FOREIGN DATA WRAPPER go_fdw
  HANDLER go_fdw_handler
  VALIDATOR go_fdw_validator;

CREATE SERVER "go-fdw"
  FOREIGN DATA WRAPPER go_fdw;
