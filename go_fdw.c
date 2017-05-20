/*-------------------------------------------------------------------------
 *
 * go_fdw.c
 * HelloWorld of foreign-data wrapper.
 *
 * written by Wataru Ikarashi <wikrsh@gmail.com>
 *
 *-------------------------------------------------------------------------
 */

#include "postgres.h"

#include <sys/stat.h>
#include <unistd.h>

#include "go_fdw.h"

PG_MODULE_MAGIC;

extern Datum go_fdw_handler(PG_FUNCTION_ARGS);
extern Datum go_fdw_validator(PG_FUNCTION_ARGS);

PG_FUNCTION_INFO_V1(go_fdw_handler);
PG_FUNCTION_INFO_V1(go_fdw_validator);

/*
 * FDW callback routines
 */

static ForeignScan *goGetForeignPlan(PlannerInfo *root,
                                        RelOptInfo *baserel,
                                        Oid foreigntableid,
                                        ForeignPath *best_path,
                                        List *tlist,
                                        List *scan_clauses,
                                        Plan *outer_plan);
//static TupleTableSlot *goIterateForeignScan(ForeignScanState *node);

Datum
go_fdw_handler(PG_FUNCTION_ARGS)
{
  GoFdwFunctions h;

  FdwRoutine *fdwroutine = makeNode(FdwRoutine);
  fdwroutine->GetForeignRelSize = goGetForeignRelSize;
  fdwroutine->GetForeignPaths = goGetForeignPaths;
  fdwroutine->GetForeignPlan = goGetForeignPlan;
  fdwroutine->ExplainForeignScan = goExplainForeignScan;
  fdwroutine->BeginForeignScan = goBeginForeignScan;
  fdwroutine->IterateForeignScan = goIterateForeignScan;
  fdwroutine->ReScanForeignScan = goReScanForeignScan;
  fdwroutine->EndForeignScan = goEndForeignScan;
  fdwroutine->AnalyzeForeignTable = goAnalyzeForeignTable;

  h.ExplainPropertyText = &ExplainPropertyText;
  h.create_foreignscan_path = &create_foreignscan_path;
  h.add_path = &add_path;
  h.BuildTupleFromCStrings = &BuildTupleFromCStrings;
  h.ExecClearTuple = &ExecClearTuple;
  h.ExecStoreTuple = &ExecStoreTuple;
  h.TupleDescGetAttInMetadata = &TupleDescGetAttInMetadata;
  goMapFuncs(h);

  PG_RETURN_POINTER(fdwroutine);
}

Datum
go_fdw_validator(PG_FUNCTION_ARGS)
{
  /* no-op */
  PG_RETURN_VOID();
}

/*
 * goGetForeignPlan
 * Create a ForeignScan plan node for scanning the foreign table
 */
static ForeignScan *
goGetForeignPlan(PlannerInfo *root,
                    RelOptInfo *baserel,
                    Oid foreigntableid,
                    ForeignPath *best_path,
                    List *tlist,
                    List *scan_clauses,
                    Plan *outer_plan)
{
  scan_clauses = extract_actual_clauses(scan_clauses, false);
  return make_foreignscan(tlist,
                          scan_clauses,
                          baserel->relid,
                          NIL,
                          best_path->fdw_private,
                          NIL,    /* no custom tlist */
                          NIL,    /* no remote quals */
                          outer_plan);
}


/*
 * goIterateForeignScan
 * Generate next record and store it into the ScanTupleSlot as a virtual tuple
 */
//static TupleTableSlot *
//goIterateForeignScan(ForeignScanState *node)
//{
//  TupleTableSlot *slot = node->ss.ss_ScanTupleSlot;
//  Relation rel;
//  AttInMetadata  *attinmeta;
//  HeapTuple tuple;
//  GoFdwExecutionState *hestate = (GoFdwExecutionState *) node->fdw_state;
//  int i;
//  int natts;
//  char **values;
//
//  if( hestate->rownum != 0 ){
//    ExecClearTuple(slot);
//    return slot;
//  }
//
//  rel = node->ss.ss_currentRelation;
//  attinmeta = TupleDescGetAttInMetadata(rel->rd_att);
//
//  natts = rel->rd_att->natts;
//  values = (char **) palloc(sizeof(char *) * natts);
//
//  for(i = 0; i < natts; i++ ){
//    values[i] = "Hello,World";
//  }
//
//  tuple = BuildTupleFromCStrings(attinmeta, values);
//  ExecStoreTuple(tuple, slot, InvalidBuffer, true);
//
//  hestate->rownum++;
//
//  return slot;
//}
