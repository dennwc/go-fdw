package main

//#cgo CFLAGS: -I/usr/include/postgresql/9.6/server -I/usr/include/postgresql/internal
//
//#include "postgres.h"
//#include "access/htup_details.h"
//#include "access/reloptions.h"
//#include "access/sysattr.h"
//#include "catalog/pg_foreign_table.h"
//#include "commands/copy.h"
//#include "commands/defrem.h"
//#include "commands/explain.h"
//#include "commands/vacuum.h"
//#include "foreign/fdwapi.h"
//#include "foreign/foreign.h"
//#include "funcapi.h"
//#include "miscadmin.h"
//#include "nodes/makefuncs.h"
//#include "optimizer/cost.h"
//#include "optimizer/pathnode.h"
//#include "optimizer/planmain.h"
//#include "optimizer/restrictinfo.h"
//#include "optimizer/var.h"
//#include "utils/memutils.h"
//#include "utils/rel.h"
//
//typedef void (*ExplainPropertyTextFunc) (const char *qlabel, const char *value, ExplainState *es);
//typedef void (*add_path_func) (RelOptInfo *parent_rel, Path *new_path);
//typedef ForeignPath* (*create_foreignscan_path_Func) (PlannerInfo *root, RelOptInfo *rel, PathTarget *target,
//    double rows, Cost startup_cost, Cost total_cost, List *pathkeys,
//    Relids required_outer, Path *fdw_outerpath, List *fdw_private);
//typedef TupleTableSlot* (*ExecClearTupleFunc) (TupleTableSlot *slot);
//typedef HeapTuple (*BuildTupleFromCStringsFunc) (AttInMetadata *attinmeta, char **values);
//typedef AttInMetadata* (*TupleDescGetAttInMetadataFunc) (TupleDesc tupdesc);
//typedef TupleTableSlot* (*ExecStoreTupleFunc) (HeapTuple tuple, TupleTableSlot *slot, Buffer buffer, bool shouldFree);
//typedef ForeignTable* (*GetForeignTableFunc) (Oid relid);
//typedef char* (*defGetStringFunc) (DefElem *def);
//
//typedef struct GoFdwExecutionState
//{
// uint tok;
//} GoFdwExecutionState;
//
//typedef struct GoFdwFunctions
//{
//  ExplainPropertyTextFunc ExplainPropertyText;
//  create_foreignscan_path_Func create_foreignscan_path;
//  add_path_func add_path;
//
//  ExecClearTupleFunc ExecClearTuple;
//  BuildTupleFromCStringsFunc BuildTupleFromCStrings;
//  TupleDescGetAttInMetadataFunc TupleDescGetAttInMetadata;
//  ExecStoreTupleFunc ExecStoreTuple;
//
//  GetForeignTableFunc GetForeignTable;
//	defGetStringFunc defGetString;
//} GoFdwFunctions;
//
//static inline void callExplainPropertyText(GoFdwFunctions h, const char *qlabel, const char *value, ExplainState *es){
//  (*(h.ExplainPropertyText))(qlabel, value, es);
//}
//
//static inline void call_add_path(GoFdwFunctions h, RelOptInfo *parent_rel, Path *new_path){
//  (*(h.add_path))(parent_rel, new_path);
//}
//
//static inline ForeignPath* call_create_foreignscan_path(GoFdwFunctions h, PlannerInfo *root, RelOptInfo *rel, PathTarget *target,
//    double rows, Cost startup_cost, Cost total_cost, List *pathkeys,
//    Relids required_outer, Path *fdw_outerpath, List *fdw_private){
//  return (*(h.create_foreignscan_path))(root,rel,target,rows,startup_cost,total_cost,pathkeys,required_outer,fdw_outerpath,fdw_private);
//}
//
//static inline TupleTableSlot* callExecClearTuple(GoFdwFunctions h, TupleTableSlot* slot){
//  return (*(h.ExecClearTuple))(slot);
//}
//
//static inline HeapTuple callBuildTupleFromCStrings(GoFdwFunctions h, AttInMetadata *attinmeta, char **values){
//  return (*(h.BuildTupleFromCStrings))(attinmeta, values);
//}
//
//static inline AttInMetadata* callTupleDescGetAttInMetadata(GoFdwFunctions h, TupleDesc tupdesc){
//  return (*(h.TupleDescGetAttInMetadata))(tupdesc);
//}
//
//static inline TupleTableSlot* callExecStoreTuple(GoFdwFunctions h, HeapTuple tuple, TupleTableSlot *slot, Buffer buffer, bool shouldFree){
//  return (*(h.ExecStoreTuple))(tuple, slot, buffer, shouldFree);
//}
//
//static inline ForeignTable* callGetForeignTable(GoFdwFunctions h, Oid relid){
//  return (*(h.GetForeignTable))(relid);
//}
//
//static inline char* callDefGetString(GoFdwFunctions h, DefElem *def){
//  return (*(h.defGetString))(def);
//}
//
//static inline GoFdwExecutionState* makeState(){
//  GoFdwExecutionState *s = (GoFdwExecutionState *) malloc(sizeof(GoFdwExecutionState));
//  return s;
//}
//
//static inline DefElem* cellGetDef(ListCell *n) { return (DefElem*)n->data.ptr_value; }
//
//static inline void freeState(GoFdwExecutionState * s){ if (s) free(s); }
import "C"

import (
	"bytes"
	"fmt"
	"sync"
	"unsafe"
)

var table Table

// SetTable sets a Go objects that will receive all FDW requests.
// Should be called with user-defined implementation in init().
func SetTable(t Table) { table = t }

// Explainer is an helper build an EXPLAIN response.
type Explainer struct {
	es *C.ExplainState
}

// Property adds a key-value property to results of EXPLAIN query.
func (e Explainer) Property(k, v string) {
	explainPropertyText(C.CString(k), C.CString(v), e.es)
}

// Options is a set of FDW options provided by user during table creation.
type Options map[string]string

// Table is a main interface for FDW table.
//
// If there are multiple tables created with this module, they can be identified by table options.
type Table interface {
	// Stats returns stats for a table.
	Stats(opts Options) TableStats
	// Scan starts a new scan of the table.
	// Iterator should not load data instantly, since Scan will be called for EXPLAIN as well.
	// Results should be fetched during Next calls.
	Scan(rel *Relation, opts Options) Iterator
}

// Iterator is an interface for table scanner implementations.
type Iterator interface {
	// Next returns next row (tuple). Nil slice means there is no more rows to scan.
	Next() []interface{}
	// Reset restarts an iterator from the beginning (possible with a new data snapshot).
	Reset()
	// Close stops an iteration and frees any resources.
	Close() error
}

// Explainable is an optional interface for Iterator that can explain it's execution plan.
type Explainable interface {
	// Explain is called during EXPLAIN query.
	Explain(e Explainer)
}

type Relation struct {
	ID      Oid
	IsValid bool
	Attr    *TupleDesc
}

type TupleDesc struct {
	TypeID  Oid
	TypeMod int
	HasOid  bool
	Attrs   []Attr // columns
}

type Attr struct {
	Name    string
	Type    Oid
	NotNull bool
}

type TableStats struct {
	Rows      uint // an estimated number of rows
	StartCost Cost
	TotalCost Cost
}

// Cost is a approximate cost of an operation. See Postgres docs for details.
type Cost float64

// Oid is a Postgres internal object ID.
type Oid uint

// A list of constants for Postgres data types
const (
	TypeBool        = Oid(C.BOOLOID)
	TypeBytes       = Oid(C.BYTEAOID)
	TypeChar        = Oid(C.CHAROID)
	TypeName        = Oid(C.NAMEOID)
	TypeInt64       = Oid(C.INT8OID)
	TypeInt16       = Oid(C.INT2OID)
	TypeInt16Vector = Oid(C.INT2VECTOROID)
	TypeInt32       = Oid(C.INT4OID)
	TypeRegProc     = Oid(C.REGPROCOID)
	TypeText        = Oid(C.TEXTOID)
	TypeOid         = Oid(C.OIDOID)

	TypeJson = Oid(C.JSONOID)
	TypeXml  = Oid(C.XMLOID)

	TypeFloat32 = Oid(C.FLOAT4OID)
	TypeFloat64 = Oid(C.FLOAT8OID)

	TypeTimestamp = Oid(C.TIMESTAMPOID)
	TypeInterval  = Oid(C.INTERVALOID)
)

// FIXME: some black magic here; we save pointers to all necessary functions (passed by glue C code)
// FIXME: it would be better to link to pg executable properly

var (
	fmu                   sync.Mutex
	explainPropertyText   func(qlabel, value *C.char, es *C.ExplainState)
	createForeignscanPath func(root *C.PlannerInfo, rel *C.RelOptInfo, target *C.PathTarget,
		rows C.double, startup_cost Cost, total_cost Cost, pathkeys *C.List,
		required_outer C.Relids, fdw_outerpath *C.Path, fdw_private *C.List) *C.ForeignPath
	addPath func(parent_rel *C.RelOptInfo, new_path *C.Path)

	execClearTuple            func(slot *C.TupleTableSlot) *C.TupleTableSlot
	buildTupleFromCStrings    func(attinmeta *C.AttInMetadata, values **C.char) C.HeapTuple
	tupleDescGetAttInMetadata func(tupdesc C.TupleDesc) *C.AttInMetadata
	execStoreTuple            func(tuple C.HeapTuple, slot *C.TupleTableSlot, buffer C.Buffer, shouldFree C.bool) *C.TupleTableSlot

	getForeignTable func(relid C.Oid) *C.ForeignTable
	defGetString    func(def *C.DefElem) *C.char
)

//export goMapFuncs
func goMapFuncs(h C.GoFdwFunctions) {
	// called the first time extension is loaded and sets all pointers to external C functions we use
	fmu.Lock()
	defer fmu.Unlock()

	explainPropertyText = func(qlabel, value *C.char, es *C.ExplainState) {
		C.callExplainPropertyText(h, qlabel, value, es)
	}
	createForeignscanPath = func(root *C.PlannerInfo, rel *C.RelOptInfo, target *C.PathTarget,
		rows C.double, startup_cost, total_cost Cost, pathkeys *C.List,
		required_outer C.Relids, fdw_outerpath *C.Path, fdw_private *C.List) *C.ForeignPath {
		return C.call_create_foreignscan_path(h,
			root, rel, target, C.double(rows),
			C.Cost(startup_cost), C.Cost(total_cost),
			pathkeys, required_outer, fdw_outerpath, fdw_private,
		)
	}
	addPath = func(parent_rel *C.RelOptInfo, new_path *C.Path) {
		C.call_add_path(h, parent_rel, new_path)
	}
	execClearTuple = func(slot *C.TupleTableSlot) *C.TupleTableSlot {
		return C.callExecClearTuple(h, slot)
	}
	buildTupleFromCStrings = func(attinmeta *C.AttInMetadata, values **C.char) C.HeapTuple {
		return C.callBuildTupleFromCStrings(h, attinmeta, values)
	}
	tupleDescGetAttInMetadata = func(tupdesc C.TupleDesc) *C.AttInMetadata {
		return C.callTupleDescGetAttInMetadata(h, tupdesc)
	}
	execStoreTuple = func(tuple C.HeapTuple, slot *C.TupleTableSlot, buffer C.Buffer, shouldFree C.bool) *C.TupleTableSlot {
		return C.callExecStoreTuple(h, tuple, slot, buffer, shouldFree)
	}
	getForeignTable = func(relid C.Oid) *C.ForeignTable {
		return C.callGetForeignTable(h, relid)
	}
	defGetString = func(def *C.DefElem) *C.char {
		return C.callDefGetString(h, def)
	}
}

//export goAnalyzeForeignTable
func goAnalyzeForeignTable(relation C.Relation, fnc *C.AcquireSampleRowsFunc, totalpages *C.BlockNumber) C.bool {
	*totalpages = 1
	return 1
}

//export goGetForeignRelSize
func goGetForeignRelSize(root *C.PlannerInfo, baserel *C.RelOptInfo, foreigntableid C.Oid) {
	// Obtain relation size estimates for a foreign table
	opts := getFTableOptions(Oid(foreigntableid))
	st := table.Stats(opts)
	baserel.rows = C.double(st.Rows)
	baserel.fdw_private = nil
}

//export goExplainForeignScan
func goExplainForeignScan(node *C.ForeignScanState, es *C.ExplainState) {
	s := getState(node.fdw_state)
	if s == nil {
		return
	}
	// Produce extra output for EXPLAIN
	if e, ok := s.Iter.(Explainable); ok {
		e.Explain(Explainer{es: es})
	}

	cs := (*C.GoFdwExecutionState)(node.fdw_state)
	clearState(uint64(cs.tok))
	C.freeState(cs)
	node.fdw_state = nil
}

//export goGetForeignPaths
func goGetForeignPaths(root *C.PlannerInfo, baserel *C.RelOptInfo, foreigntableid C.Oid) {
	// Create possible access paths for a scan on the foreign table
	opts := getFTableOptions(Oid(foreigntableid))
	st := table.Stats(opts)
	addPath(baserel,
		(*C.Path)(unsafe.Pointer(createForeignscanPath(
			root,
			baserel,
			baserel.reltarget,
			baserel.rows,
			st.StartCost,
			st.TotalCost,
			nil, // no pathkeys
			nil, // no outer rel either
			nil, // no extra plan
			nil,
		))),
	)
}

//export goBeginForeignScan
func goBeginForeignScan(node *C.ForeignScanState, eflags C.int) {
	rel := buildRelation(node.ss.ss_currentRelation)
	opts := getFTableOptions(rel.ID)
	s := &State{
		Rel: rel, Opts: opts,
		Iter: table.Scan(rel, opts),
	}
	i := saveState(s)

	//plan := node.ss.ps.plan
	//plan := (*C.ForeignScan)(unsafe.Pointer(node.ss.ps.plan))

	cs := C.makeState()
	cs.tok = C.uint(i)
	node.fdw_state = unsafe.Pointer(cs)

	//if eflags&C.EXEC_FLAG_EXPLAIN_ONLY != 0 {
	//	return // Do nothing in EXPLAIN
	//}
}

//export goIterateForeignScan
func goIterateForeignScan(node *C.ForeignScanState) *C.TupleTableSlot {
	s := getState(node.fdw_state)

	slot := node.ss.ss_ScanTupleSlot
	execClearTuple(slot)

	row := s.Iter.Next()
	if row == nil {
		return slot
	}

	values := make([]*C.char, len(s.Rel.Attr.Attrs))
	for i := range s.Rel.Attr.Attrs {
		v := row[i]
		if v == nil {
			p := unsafe.Pointer(uintptr(unsafe.Pointer(slot.tts_isnull)) + uintptr(C.sizeof_bool))
			*((*C.bool)(p)) = 1
			continue
		}
		values[i] = C.CString(fmt.Sprint(v))
	}

	rel := node.ss.ss_currentRelation
	attinmeta := tupleDescGetAttInMetadata(rel.rd_att)
	tuple := buildTupleFromCStrings(attinmeta, (**C.char)(&values[0]))
	execStoreTuple(tuple, slot, C.InvalidBuffer, 1)
	return slot
}

//export goReScanForeignScan
func goReScanForeignScan(node *C.ForeignScanState) {
	// Rescan table, possibly with new parameters
	s := getState(node.fdw_state)
	s.Iter.Reset()
}

//export goEndForeignScan
func goEndForeignScan(node *C.ForeignScanState) {
	// Finish scanning foreign table and dispose objects used for this scan
	s := getState(node.fdw_state)
	if s == nil {
		return
	}
	cs := (*C.GoFdwExecutionState)(node.fdw_state)
	clearState(uint64(cs.tok))
	C.freeState(cs)
	node.fdw_state = nil
}

type State struct {
	Rel  *Relation
	Opts map[string]string
	Iter Iterator
}

var (
	mu   sync.RWMutex
	si   uint64
	sess = make(map[uint64]*State)
)

func saveState(s *State) uint64 {
	mu.Lock()
	si++
	i := si
	sess[i] = s
	mu.Unlock()
	return i
}

func clearState(i uint64) {
	mu.Lock()
	delete(sess, i)
	mu.Unlock()
}

func getState(p unsafe.Pointer) *State {
	if p == nil {
		return nil
	}
	cs := (*C.GoFdwExecutionState)(p)
	i := uint64(cs.tok)
	mu.RLock()
	s := sess[i]
	mu.RUnlock()
	return s
}

func getFTableOptions(id Oid) Options {
	f := getForeignTable(C.Oid(id))
	return getOptions(f.options)
}

func getOptions(opts *C.List) Options {
	m := make(Options)
	for it := opts.head; it != nil; it = it.next {
		el := C.cellGetDef(it)
		name := C.GoString(el.defname)
		val := C.GoString(defGetString(el))
		m[name] = val
	}
	return m
}

func buildRelation(rel C.Relation) *Relation {
	r := &Relation{
		ID:      Oid(rel.rd_id),
		IsValid: goBool(rel.rd_isvalid),
		Attr:    buildTupleDesc(rel.rd_att),
	}
	return r
}

func goBool(b C.bool) bool {
	return b != 0
}

func goString(p unsafe.Pointer, n int) string {
	b := C.GoBytes(p, C.int(n))
	i := bytes.IndexByte(b, 0)
	if i < 0 {
		i = len(b)
	}
	return string(b[:i])
}

func buildTupleDesc(desc C.TupleDesc) *TupleDesc {
	if desc == nil {
		return nil
	}
	d := &TupleDesc{
		TypeID:  Oid(desc.tdtypeid),
		TypeMod: int(desc.tdtypmod),
		HasOid:  goBool(desc.tdhasoid),
		Attrs:   make([]Attr, 0, int(desc.natts)),
	}
	for i := 0; i < cap(d.Attrs); i++ {
		off := uintptr(i) * uintptr(C.sizeof_Form_pg_attribute)
		p := *(*C.Form_pg_attribute)(unsafe.Pointer(uintptr(unsafe.Pointer(desc.attrs)) + off))
		d.Attrs = append(d.Attrs, buildAttr(p))
	}
	return d
}

const nameLen = C.NAMEDATALEN

func buildAttr(attr *C.FormData_pg_attribute) (out Attr) {
	out.Name = goString(unsafe.Pointer(&attr.attname.data[0]), nameLen)
	out.Type = Oid(attr.atttypid)
	out.NotNull = goBool(attr.attnotnull)
	return
}

// required by buildmode=c-archive

func main() {}
