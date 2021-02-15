/*

Copyright (c) 2021 - Present. Blend Labs, Inc. All rights reserved
Use of this source code is governed by a MIT license that can be found in the LICENSE file.

*/

package db

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/blend/go-sdk/bufferutil"
	"github.com/blend/go-sdk/ex"
	"github.com/blend/go-sdk/logger"
)

// Invocation is a specific operation against a context.
type Invocation struct {
	DB DB

	/* invocation state */
	Label string

	/* context */
	Context context.Context
	Cancel  func()

	/* dependencies */
	Config     Config
	Log        logger.Triggerable
	BufferPool *bufferutil.Pool

	/* logging hooks */
	StatementInterceptor StatementInterceptor
	Tracer               Tracer
	StartTime            time.Time
	TraceFinisher        TraceFinisher
}

// Exec executes a sql statement with a given set of arguments and returns the rows affected.
func (i *Invocation) Exec(statement string, args ...interface{}) (res sql.Result, err error) {
	statement = i.start(statement)
	defer func() { err = i.finish(statement, recover(), res, err) }()

	res, err = i.DB.ExecContext(i.Context, statement, args...)
	if err != nil {
		err = Error(err)
		return
	}
	return
}

// Query returns a new query object for a given sql query and arguments.
func (i *Invocation) Query(statement string, args ...interface{}) *Query {
	return &Query{
		Invocation: i,
		Statement:  i.start(statement),
		Args:       args,
	}
}

// Get returns a given object based on a group of primary key ids within a transaction.
func (i *Invocation) Get(object DatabaseMapped, ids ...interface{}) (found bool, err error) {
	if len(ids) == 0 {
		err = Error(ErrInvalidIDs)
		return
	}

	var queryBody string
	if i.Label, queryBody, err = i.generateGet(object); err != nil {
		err = Error(err)
		return
	}
	return i.Query(queryBody, ids...).Out(object)
}

// All returns all rows of an object mapped table wrapped in a transaction.
func (i *Invocation) All(collection interface{}) (err error) {
	var queryBody string
	i.Label, queryBody = i.generateGetAll(collection)
	return i.Query(queryBody).OutMany(collection)
}

// Create writes an object to the database within a transaction.
func (i *Invocation) Create(object DatabaseMapped) (err error) {
	var queryBody string
	var insertCols, autos *ColumnCollection
	var res sql.Result
	defer func() { err = i.finish(queryBody, recover(), res, err) }()

	i.Label, queryBody, insertCols, autos = i.generateCreate(object)

	queryBody = i.start(queryBody)
	if autos.Len() == 0 {
		if res, err = i.DB.ExecContext(i.Context, queryBody, insertCols.ColumnValues(object)...); err != nil {
			err = Error(err)
			return
		}
		return
	}

	autoValues := i.autoValues(autos)
	if err = i.DB.QueryRowContext(i.Context, queryBody, insertCols.ColumnValues(object)...).Scan(autoValues...); err != nil {
		err = Error(err)
		return
	}
	if err = i.setAutos(object, autos, autoValues); err != nil {
		err = Error(err)
		return
	}

	return
}

// CreateIfNotExists writes an object to the database if it does not already exist within a transaction.
// This will _ignore_ auto columns, as they will always invalidate the assertion that there already exists
// a row with a given primary key set.
func (i *Invocation) CreateIfNotExists(object DatabaseMapped) (err error) {
	var queryBody string
	var insertCols *ColumnCollection
	var res sql.Result
	defer func() { err = i.finish(queryBody, recover(), res, err) }()

	i.Label, queryBody, insertCols = i.generateCreateIfNotExists(object)

	queryBody = i.start(queryBody)
	if res, err = i.DB.ExecContext(i.Context, queryBody, insertCols.ColumnValues(object)...); err != nil {
		err = Error(err)
	}
	return
}

// CreateMany writes many objects to the database in a single insert.
func (i *Invocation) CreateMany(objects interface{}) (err error) {
	return i.insertOrUpsertMany(objects, false)
}

// UpsertMany writes many objects to the database in a single upsert.
func (i *Invocation) UpsertMany(objects interface{}) (err error) {
	return i.insertOrUpsertMany(objects, true)
}

// insertOrUpsertManinsertOrUpsertMany writes many objects to the database in a single insert or upsert.
func (i *Invocation) insertOrUpsertMany(objects interface{}, overwrite bool) (err error) {
	var queryBody string
	var insertCols *ColumnCollection
	var sliceValue reflect.Value
	var res sql.Result
	defer func() { err = i.finish(queryBody, recover(), res, err) }()

	if overwrite {
		queryBody, insertCols, sliceValue = i.generateUpsertMany(objects)
	} else {
		queryBody, insertCols, sliceValue = i.generateCreateMany(objects)
	}
	if sliceValue.Len() == 0 {
		// If there is nothing to create, then we're done here
		return
	}

	queryBody = i.start(queryBody)
	var colValues []interface{}
	for row := 0; row < sliceValue.Len(); row++ {
		colValues = append(colValues, insertCols.ColumnValues(sliceValue.Index(row).Interface())...)
	}

	res, err = i.DB.ExecContext(i.Context, queryBody, colValues...)
	if err != nil {
		err = Error(err)
		return
	}
	return
}

// Update updates an object wrapped in a transaction. Returns whether or not any rows have been updated and potentially
// an error. If ErrTooManyRows is returned, it's important to note that due to https://github.com/golang/go/issues/7898,
// the Update HAS BEEN APPLIED. Its on the developer using UPDATE to ensure his tags are correct and/or execute it in a
// transaction and roll back on this error
func (i *Invocation) Update(object DatabaseMapped) (updated bool, err error) {
	var queryBody string
	var pks, updateCols *ColumnCollection
	var res sql.Result
	defer func() { err = i.finish(queryBody, recover(), res, err) }()

	i.Label, queryBody, pks, updateCols = i.generateUpdate(object)

	queryBody = i.start(queryBody)
	res, err = i.DB.ExecContext(
		i.Context,
		queryBody,
		append(updateCols.ColumnValues(object), pks.ColumnValues(object)...)...,
	)
	if err != nil {
		err = Error(err)
		return
	}

	var rowCount int64
	rowCount, err = res.RowsAffected()
	if err != nil {
		err = Error(err)
		return
	}
	if rowCount > 0 {
		updated = true
	}
	if rowCount > 1 {
		err = Error(ErrTooManyRows)
	}
	return
}

// Upsert inserts the object if it doesn't exist already (as defined by its primary keys) or updates it atomically.
// It returns `found` as true if the effect was an upsert, i.e. the pk was found.
func (i *Invocation) Upsert(object DatabaseMapped) (err error) {
	var queryBody string
	var autos, upsertCols *ColumnCollection
	defer func() { err = i.finish(queryBody, recover(), nil, err) }()

	i.Label, queryBody, autos, upsertCols = i.generateUpsert(object)

	queryBody = i.start(queryBody)
	if autos.Len() == 0 {
		if _, err = i.Exec(queryBody, upsertCols.ColumnValues(object)...); err != nil {
			return
		}
		return
	}

	autoValues := i.autoValues(autos)
	if err = i.DB.QueryRowContext(i.Context, queryBody, upsertCols.ColumnValues(object)...).Scan(autoValues...); err != nil {
		err = Error(err)
		return
	}
	if err = i.setAutos(object, autos, autoValues); err != nil {
		err = Error(err)
		return
	}
	return
}

// Exists returns a bool if a given object exists (utilizing the primary key columns if they exist) wrapped in a transaction.
func (i *Invocation) Exists(object DatabaseMapped) (exists bool, err error) {
	var queryBody string
	var pks *ColumnCollection
	defer func() { err = i.finish(queryBody, recover(), nil, err) }()

	if i.Label, queryBody, pks, err = i.generateExists(object); err != nil {
		err = Error(err)
		return
	}
	queryBody = i.start(queryBody)
	var value int
	if queryErr := i.DB.QueryRowContext(i.Context, queryBody, pks.ColumnValues(object)...).Scan(&value); queryErr != nil && !ex.Is(queryErr, sql.ErrNoRows) {
		err = Error(queryErr)
		return
	}
	exists = value != 0
	return
}

// Delete deletes an object from the database wrapped in a transaction. Returns whether or not any rows have been deleted
// and potentially an error. If ErrTooManyRows is returned, it's important to note that due to
// https://github.com/golang/go/issues/7898, the Delete HAS BEEN APPLIED on the current transaction. Its on the
// developer using Delete to ensure their tags are correct and/or ensure theit Tx rolls back on this error.
func (i *Invocation) Delete(object DatabaseMapped) (deleted bool, err error) {
	var queryBody string
	var pks *ColumnCollection
	var res sql.Result
	defer func() { err = i.finish(queryBody, recover(), res, err) }()

	if i.Label, queryBody, pks, err = i.generateDelete(object); err != nil {
		return
	}

	queryBody = i.start(queryBody)
	res, err = i.DB.ExecContext(i.Context, queryBody, pks.ColumnValues(object)...)
	if err != nil {
		err = Error(err)
		return
	}

	var rowCount int64
	rowCount, err = res.RowsAffected()
	if err != nil {
		err = Error(err)
		return
	}
	if rowCount > 0 {
		deleted = true
	}
	if rowCount > 1 {
		err = Error(ErrTooManyRows)
	}
	return
}

// --------------------------------------------------------------------------------
// query body generators
// --------------------------------------------------------------------------------

func (i *Invocation) generateGet(object DatabaseMapped) (cachePlan, queryBody string, err error) {
	tableName := TableName(object)

	cols := Columns(object).NotReadOnly()
	pks := cols.PrimaryKeys()
	if pks.Len() == 0 {
		err = Error(ErrNoPrimaryKey)
		return
	}

	queryBodyBuffer := i.BufferPool.Get()
	defer i.BufferPool.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("SELECT ")
	for i, name := range cols.ColumnNames() {
		queryBodyBuffer.WriteString(name)
		if i < (cols.Len() - 1) {
			queryBodyBuffer.WriteRune(',')
		}
	}

	queryBodyBuffer.WriteString(" FROM ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" WHERE ")

	for i, pk := range pks.Columns() {
		queryBodyBuffer.WriteString(pk.ColumnName)
		queryBodyBuffer.WriteString(" = ")
		queryBodyBuffer.WriteString("$" + strconv.Itoa(i+1))

		if i < (pks.Len() - 1) {
			queryBodyBuffer.WriteString(" AND ")
		}
	}

	cachePlan = fmt.Sprintf("%s_get", tableName)
	queryBody = queryBodyBuffer.String()
	return
}

func (i *Invocation) generateGetAll(collection interface{}) (statementLabel, queryBody string) {
	collectionType := ReflectSliceType(collection)
	tableName := TableNameByType(collectionType)

	cols := ColumnsFromType(tableName, ReflectSliceType(collection)).NotReadOnly()

	queryBodyBuffer := i.BufferPool.Get()
	defer i.BufferPool.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("SELECT ")
	for i, name := range cols.ColumnNames() {
		queryBodyBuffer.WriteString(name)
		if i < (cols.Len() - 1) {
			queryBodyBuffer.WriteRune(',')
		}
	}
	queryBodyBuffer.WriteString(" FROM ")
	queryBodyBuffer.WriteString(tableName)

	queryBody = queryBodyBuffer.String()
	statementLabel = tableName + "_get_all"
	return
}

func (i *Invocation) generateCreate(object DatabaseMapped) (statementLabel, queryBody string, insertCols, autos *ColumnCollection) {
	tableName := TableName(object)

	cols := Columns(object)
	insertCols = cols.InsertColumns().ConcatWith(cols.Autos().NotZero(object))
	autos = cols.Autos()

	queryBodyBuffer := i.BufferPool.Get()
	defer i.BufferPool.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("INSERT INTO ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" (")
	for i, name := range insertCols.ColumnNames() {
		queryBodyBuffer.WriteString(name)
		if i < (insertCols.Len() - 1) {
			queryBodyBuffer.WriteRune(',')
		}
	}
	queryBodyBuffer.WriteString(") VALUES (")
	for x := 0; x < insertCols.Len(); x++ {
		queryBodyBuffer.WriteString("$" + strconv.Itoa(x+1))
		if x < (insertCols.Len() - 1) {
			queryBodyBuffer.WriteRune(',')
		}
	}
	queryBodyBuffer.WriteString(")")

	if autos.Len() > 0 {
		queryBodyBuffer.WriteString(" RETURNING ")
		queryBodyBuffer.WriteString(autos.ColumnNamesCSV())
	}

	queryBody = queryBodyBuffer.String()
	statementLabel = tableName + "_create"
	return
}

func (i *Invocation) generateCreateIfNotExists(object DatabaseMapped) (statementLabel, queryBody string, insertCols *ColumnCollection) {
	cols := Columns(object)

	insertCols = cols.InsertColumns().ConcatWith(cols.Autos().NotZero(object))

	pks := cols.PrimaryKeys()
	tableName := TableName(object)

	queryBodyBuffer := i.BufferPool.Get()
	defer i.BufferPool.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("INSERT INTO ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" (")
	for i, name := range insertCols.ColumnNames() {
		queryBodyBuffer.WriteString(name)
		if i < (insertCols.Len() - 1) {
			queryBodyBuffer.WriteRune(',')
		}
	}
	queryBodyBuffer.WriteString(") VALUES (")
	for x := 0; x < insertCols.Len(); x++ {
		queryBodyBuffer.WriteString("$" + strconv.Itoa(x+1))
		if x < (insertCols.Len() - 1) {
			queryBodyBuffer.WriteRune(',')
		}
	}
	queryBodyBuffer.WriteString(")")

	if pks.Len() > 0 {
		queryBodyBuffer.WriteString(" ON CONFLICT (")
		pkColumnNames := pks.ColumnNames()
		for i, name := range pkColumnNames {
			queryBodyBuffer.WriteString(name)
			if i < len(pkColumnNames)-1 {
				queryBodyBuffer.WriteRune(',')
			}
		}
		queryBodyBuffer.WriteString(") DO NOTHING")
	}

	queryBody = queryBodyBuffer.String()
	statementLabel = tableName + "_create_if_not_exists"
	return
}

func (i *Invocation) generateUpsertMany(objects interface{}) (queryBody string, insertCols *ColumnCollection, sliceValue reflect.Value) {
	queryBodyInsertMany, insertCols, sliceValue := i.generateCreateMany(objects)

	queryBodyBuffer := i.BufferPool.Get()
	defer i.BufferPool.Put(queryBodyBuffer)
	queryBodyBuffer.WriteString(queryBodyInsertMany)

	uks := insertCols.UniqueKeys()
	if uks.Len() > 0 {
		queryBodyBuffer.WriteString(" ON CONFLICT (")
		ukColumnNames := uks.ColumnNames()
		for i, name := range ukColumnNames {
			queryBodyBuffer.WriteString(name)
			if i < len(ukColumnNames)-1 {
				queryBodyBuffer.WriteRune(',')
			}
		}
		queryBodyBuffer.WriteString(") DO UPDATE SET ")

		for i, name := range insertCols.ColumnNames() {
			queryBodyBuffer.WriteString(fmt.Sprintf("%s=Excluded.%s", name, name))
			if i < (insertCols.Len() - 1) {
				queryBodyBuffer.WriteRune(',')
			}
		}
	}
	queryBody = queryBodyBuffer.String()
	return
}

func (i *Invocation) generateCreateMany(objects interface{}) (queryBody string, insertCols *ColumnCollection, sliceValue reflect.Value) {
	sliceValue = ReflectValue(objects)
	sliceType := ReflectSliceType(objects)
	tableName := TableNameByType(sliceType)

	cols := ColumnsFromType(tableName, sliceType)
	insertCols = cols.InsertColumns()

	queryBodyBuffer := i.BufferPool.Get()
	defer i.BufferPool.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("INSERT INTO ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" (")
	for i, name := range insertCols.ColumnNames() {
		queryBodyBuffer.WriteString(name)
		if i < (insertCols.Len() - 1) {
			queryBodyBuffer.WriteRune(',')
		}
	}

	queryBodyBuffer.WriteString(") VALUES ")

	metaIndex := 1
	for x := 0; x < sliceValue.Len(); x++ {
		queryBodyBuffer.WriteString("(")
		for y := 0; y < insertCols.Len(); y++ {
			queryBodyBuffer.WriteString(fmt.Sprintf("$%d", metaIndex))
			metaIndex = metaIndex + 1
			if y < insertCols.Len()-1 {
				queryBodyBuffer.WriteRune(',')
			}
		}
		queryBodyBuffer.WriteString(")")
		if x < sliceValue.Len()-1 {
			queryBodyBuffer.WriteRune(',')
		}
	}

	queryBody = queryBodyBuffer.String()
	return
}

func (i *Invocation) generateUpdate(object DatabaseMapped) (statementLabel, queryBody string, pks, updateCols *ColumnCollection) {
	tableName := TableName(object)

	cols := Columns(object)

	pks = cols.PrimaryKeys()
	updateCols = cols.UpdateColumns()

	queryBodyBuffer := i.BufferPool.Get()
	defer i.BufferPool.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("UPDATE ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" SET ")

	var updateColIndex int
	var col Column
	for ; updateColIndex < updateCols.Len(); updateColIndex++ {
		col = updateCols.Columns()[updateColIndex]
		queryBodyBuffer.WriteString(col.ColumnName)
		queryBodyBuffer.WriteString(" = $" + strconv.Itoa(updateColIndex+1))
		if updateColIndex != (updateCols.Len() - 1) {
			queryBodyBuffer.WriteRune(',')
		}
	}

	queryBodyBuffer.WriteString(" WHERE ")
	for i, pk := range pks.Columns() {
		queryBodyBuffer.WriteString(pk.ColumnName)
		queryBodyBuffer.WriteString(" = ")
		queryBodyBuffer.WriteString("$" + strconv.Itoa(i+(updateColIndex+1)))

		if i < (pks.Len() - 1) {
			queryBodyBuffer.WriteString(" AND ")
		}
	}

	queryBody = queryBodyBuffer.String()
	statementLabel = tableName + "_update"
	return
}

func (i *Invocation) generateUpsert(object DatabaseMapped) (statementLabel, queryBody string, autos, insertCols *ColumnCollection) {
	tableName := TableName(object)
	cols := Columns(object)
	updates := cols.UpdateColumns()
	updateCols := updates.Columns()

	// these should be not auto columns, and columns that are auto and are set
	insertCols = cols.InsertColumns().ConcatWith(cols.Autos().NotZero(object))

	insertColNames := insertCols.ColumnNames()

	// autos are read out on insert (but only if unset)
	autos = cols.Autos().Zero(object)
	pks := cols.PrimaryKeys()
	pkNames := pks.ColumnNames()

	queryBodyBuffer := i.BufferPool.Get()
	defer i.BufferPool.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("INSERT INTO ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" (")
	for i, name := range insertColNames {
		queryBodyBuffer.WriteString(name)
		if i < len(insertColNames)-1 {
			queryBodyBuffer.WriteRune(',')
		}
	}

	queryBodyBuffer.WriteString(") VALUES (")

	for x := 0; x < insertCols.Len(); x++ {
		queryBodyBuffer.WriteString("$" + strconv.Itoa(x+1))
		if x < (insertCols.Len() - 1) {
			queryBodyBuffer.WriteRune(',')
		}
	}

	queryBodyBuffer.WriteString(")")

	if pks.Len() > 0 {
		tokenMap := map[string]string{}
		for i, col := range insertCols.Columns() {
			tokenMap[col.ColumnName] = "$" + strconv.Itoa(i+1)
		}

		queryBodyBuffer.WriteString(" ON CONFLICT (")

		for i, name := range pkNames {
			queryBodyBuffer.WriteString(name)
			if i < len(pkNames)-1 {
				queryBodyBuffer.WriteRune(',')
			}
		}
		queryBodyBuffer.WriteString(") DO UPDATE SET ")

		for i, col := range updateCols {
			queryBodyBuffer.WriteString(col.ColumnName + " = " + tokenMap[col.ColumnName])
			if i < (len(updateCols) - 1) {
				queryBodyBuffer.WriteRune(',')
			}
		}
	}
	if autos.Len() > 0 {
		queryBodyBuffer.WriteString(" RETURNING ")
		queryBodyBuffer.WriteString(autos.ColumnNamesCSV())
	}

	queryBody = queryBodyBuffer.String()
	statementLabel = tableName + "_upsert"
	return
}

func (i *Invocation) generateExists(object DatabaseMapped) (statementLabel, queryBody string, pks *ColumnCollection, err error) {
	tableName := TableName(object)
	pks = Columns(object).PrimaryKeys()
	if pks.Len() == 0 {
		err = Error(ErrNoPrimaryKey)
		return
	}
	queryBodyBuffer := i.BufferPool.Get()
	defer i.BufferPool.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("SELECT 1 FROM ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" WHERE ")
	for i, pk := range pks.Columns() {
		queryBodyBuffer.WriteString(pk.ColumnName)
		queryBodyBuffer.WriteString(" = ")
		queryBodyBuffer.WriteString("$" + strconv.Itoa(i+1))

		if i < (pks.Len() - 1) {
			queryBodyBuffer.WriteString(" AND ")
		}
	}
	statementLabel = tableName + "_exists"
	queryBody = queryBodyBuffer.String()
	return
}

func (i *Invocation) generateDelete(object DatabaseMapped) (statementLabel, queryBody string, pks *ColumnCollection, err error) {
	tableName := TableName(object)
	pks = Columns(object).PrimaryKeys()
	if len(pks.Columns()) == 0 {
		err = Error(ErrNoPrimaryKey)
		return
	}
	queryBodyBuffer := i.BufferPool.Get()
	defer i.BufferPool.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("DELETE FROM ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" WHERE ")
	for i, pk := range pks.Columns() {
		queryBodyBuffer.WriteString(pk.ColumnName)
		queryBodyBuffer.WriteString(" = ")
		queryBodyBuffer.WriteString("$" + strconv.Itoa(i+1))

		if i < (pks.Len() - 1) {
			queryBodyBuffer.WriteString(" AND ")
		}
	}
	statementLabel = tableName + "_delete"
	queryBody = queryBodyBuffer.String()
	return
}

// --------------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------------

// autoValues returns references to the auto updatd fields for a given column collection.
func (i *Invocation) autoValues(autos *ColumnCollection) []interface{} {
	autoValues := make([]interface{}, autos.Len())
	for i, autoCol := range autos.Columns() {
		autoValues[i] = reflect.New(reflect.PtrTo(autoCol.FieldType)).Interface()
	}
	return autoValues
}

// setAutos sets the automatic values for a given object.
func (i *Invocation) setAutos(object DatabaseMapped, autos *ColumnCollection, autoValues []interface{}) (err error) {
	for index := 0; index < len(autoValues); index++ {
		err = autos.Columns()[index].SetValue(object, autoValues[index])
		if err != nil {
			err = Error(err)
			return
		}
	}
	return
}

// start runs on start steps.
func (i *Invocation) start(statement string) string {
	i.StartTime = time.Now()
	if i.StatementInterceptor != nil {
		statement = i.StatementInterceptor(i.Label, statement)
	}
	if i.Tracer != nil && !IsSkipQueryLogging(i.Context) {
		i.TraceFinisher = i.Tracer.Query(i.Context, i.Config, i.Label, statement)
	}
	return statement
}

// finish runs on complete steps.
func (i *Invocation) finish(statement string, r interface{}, res sql.Result, err error) error {
	if i.Cancel != nil {
		i.Cancel()
	}
	if r != nil {
		err = ex.Nest(err, ex.New(r))
	}
	if i.Log != nil && !IsSkipQueryLogging(i.Context) {
		qe := NewQueryEvent(statement, time.Now().UTC().Sub(i.StartTime))
		qe.Username = i.Config.Username
		qe.Database = i.Config.DatabaseOrDefault()
		qe.Label = i.Label
		qe.Engine = i.Config.EngineOrDefault()
		qe.Err = err
		i.Log.TriggerContext(i.Context, qe)
	}
	if i.TraceFinisher != nil && !IsSkipQueryLogging(i.Context) {
		i.TraceFinisher.FinishQuery(i.Context, res, err)
	}
	if err != nil {
		err = Error(err, ex.OptMessage(statement))
	}
	return err
}
