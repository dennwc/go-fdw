package main

//#cgo CFLAGS: -I/usr/include/postgresql/9.6/server -I/usr/include/postgresql/internal
//
//#include <sys/stat.h>
//#include <unistd.h>
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
//static inline GoFdwExecutionState* makeState(){
//  GoFdwExecutionState *s = (GoFdwExecutionState *) malloc(sizeof(GoFdwExecutionState));
//  return s;
//}
//
//static inline void freeState(GoFdwExecutionState * s){ if (s) free(s); }
import "C"

import (
	"fmt"
	"log"
	"sync"
	"unsafe"
)

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
)

type Cost float64

//export goMapFuncs
func goMapFuncs(h C.GoFdwFunctions) {
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
}

const rows = 3

//export goGetForeignRelSize
func goGetForeignRelSize(root *C.PlannerInfo, baserel *C.RelOptInfo, foreigntableid C.Oid) {
	// Obtain relation size estimates for a foreign table
	baserel.rows = rows
	baserel.fdw_private = nil
}

//export goAnalyzeForeignTable
func goAnalyzeForeignTable(relation C.Relation, fnc *C.AcquireSampleRowsFunc, totalpages *C.BlockNumber) C.bool {
	*totalpages = 1
	return 1
}

//export goExplainForeignScan
func goExplainForeignScan(node *C.ForeignScanState, es *C.ExplainState) {
	// Produce extra output for EXPLAIN
	explainPropertyText(C.CString("Go9"), C.CString("we are running from Go!"), es)
}

//export goGetForeignPaths
func goGetForeignPaths(root *C.PlannerInfo, baserel *C.RelOptInfo, foreigntableid C.Oid) {
	// Create Possible access paths for a scan on the foreign table
	addPath(baserel,
		(*C.Path)(unsafe.Pointer(createForeignscanPath(
			root,
			baserel,
			nil,
			baserel.rows,
			10,   // startup_cost
			1000, // total_cost
			nil,  // no pathkeys
			nil,  // no outer rel either
			nil,  // no extra plan
			nil,
		))),
	)
}

type State struct {
	Row int
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
	cs := (*C.GoFdwExecutionState)(p)
	mu.RLock()
	s := sess[uint64(cs.tok)]
	mu.RUnlock()
	return s
}

//export goBeginForeignScan
func goBeginForeignScan(node *C.ForeignScanState, eflags C.int) {
	if eflags&C.EXEC_FLAG_EXPLAIN_ONLY != 0 {
		return // Do nothing in EXPLAIN
	}
	i := saveState(&State{Row: 0})
	log.Printf("begin scan (%d): %x", i, int(eflags))
	cs := C.makeState()
	cs.tok = C.uint(i)
	node.fdw_state = unsafe.Pointer(cs)
}

//export goIterateForeignScan
func goIterateForeignScan(node *C.ForeignScanState) *C.TupleTableSlot {
	s := getState(node.fdw_state)
	log.Printf("scan (%+v)", s)

	slot := node.ss.ss_ScanTupleSlot
	execClearTuple(slot)

	if s.Row >= rows {
		log.Printf("scan end")
		return slot
	}

	rel := node.ss.ss_currentRelation
	attinmeta := tupleDescGetAttInMetadata(rel.rd_att)

	natts := int(rel.rd_att.natts)
	values := make([]*C.char, natts)

	for i := 0; i < natts; i++ {
		values[i] = C.CString(fmt.Sprintf("Row: %d, Col: %d", s.Row, i))
	}

	tuple := buildTupleFromCStrings(attinmeta, (**C.char)(&values[0]))
	execStoreTuple(tuple, slot, C.InvalidBuffer, 1)

	s.Row++
	return slot
}

//export goReScanForeignScan
func goReScanForeignScan(node *C.ForeignScanState) {
	// Rescan table, possibly with new parameters
	s := getState(node.fdw_state)
	log.Printf("reset (%+v)", s)
	s.Row = 0
}

//export goEndForeignScan
func goEndForeignScan(node *C.ForeignScanState) {
	// Finish scanning foreign table and dispose objects used for this scan
	if node.fdw_state != nil {
		cs := (*C.GoFdwExecutionState)(node.fdw_state)
		clearState(uint64(cs.tok))
		C.freeState(cs)
		node.fdw_state = nil
	}
}

func main() {}