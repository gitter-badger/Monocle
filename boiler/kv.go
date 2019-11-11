// Code generated by SQLBoiler 3.5.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package boiler

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/volatiletech/sqlboiler/queries/qmhelper"
	"github.com/volatiletech/sqlboiler/strmangle"
	"github.com/volatiletech/sqlboiler/types"
)

// KV is an object representing the database table.
type KV struct {
	K         string     `db:"k" boil:"k" json:"k" toml:"k" yaml:"k"`
	V         types.JSON `db:"v" boil:"v" json:"v" toml:"v" yaml:"v"`
	CreatedAt time.Time  `db:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`

	R *kvR `db:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L kvL  `db:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var KVColumns = struct {
	K         string
	V         string
	CreatedAt string
	UpdatedAt string
}{
	K:         "k",
	V:         "v",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// Generated where

type whereHelpertypes_JSON struct{ field string }

func (w whereHelpertypes_JSON) EQ(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.EQ, x)
}
func (w whereHelpertypes_JSON) NEQ(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelpertypes_JSON) LT(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpertypes_JSON) LTE(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpertypes_JSON) GT(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpertypes_JSON) GTE(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

var KVWhere = struct {
	K         whereHelperstring
	V         whereHelpertypes_JSON
	CreatedAt whereHelpertime_Time
	UpdatedAt whereHelpertime_Time
}{
	K:         whereHelperstring{field: "`kv`.`k`"},
	V:         whereHelpertypes_JSON{field: "`kv`.`v`"},
	CreatedAt: whereHelpertime_Time{field: "`kv`.`created_at`"},
	UpdatedAt: whereHelpertime_Time{field: "`kv`.`updated_at`"},
}

// KVRels is where relationship names are stored.
var KVRels = struct {
}{}

// kvR is where relationships are stored.
type kvR struct {
}

// NewStruct creates a new relationship struct
func (*kvR) NewStruct() *kvR {
	return &kvR{}
}

// kvL is where Load methods for each relationship are stored.
type kvL struct{}

var (
	kvAllColumns            = []string{"k", "v", "created_at", "updated_at"}
	kvColumnsWithoutDefault = []string{"k", "v", "created_at", "updated_at"}
	kvColumnsWithDefault    = []string{}
	kvPrimaryKeyColumns     = []string{"k"}
)

type (
	// KVSlice is an alias for a slice of pointers to KV.
	// This should generally be used opposed to []KV.
	KVSlice []*KV

	kvQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	kvType                 = reflect.TypeOf(&KV{})
	kvMapping              = queries.MakeStructMapping(kvType)
	kvPrimaryKeyMapping, _ = queries.BindMapping(kvType, kvMapping, kvPrimaryKeyColumns)
	kvInsertCacheMut       sync.RWMutex
	kvInsertCache          = make(map[string]insertCache)
	kvUpdateCacheMut       sync.RWMutex
	kvUpdateCache          = make(map[string]updateCache)
	kvUpsertCacheMut       sync.RWMutex
	kvUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

// OneG returns a single kv record from the query using the global executor.
func (q kvQuery) OneG(ctx context.Context) (*KV, error) {
	return q.One(ctx, boil.GetContextDB())
}

// One returns a single kv record from the query.
func (q kvQuery) One(ctx context.Context, exec boil.ContextExecutor) (*KV, error) {
	o := &KV{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for kv")
	}

	return o, nil
}

// AllG returns all KV records from the query using the global executor.
func (q kvQuery) AllG(ctx context.Context) (KVSlice, error) {
	return q.All(ctx, boil.GetContextDB())
}

// All returns all KV records from the query.
func (q kvQuery) All(ctx context.Context, exec boil.ContextExecutor) (KVSlice, error) {
	var o []*KV

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to KV slice")
	}

	return o, nil
}

// CountG returns the count of all KV records in the query, and panics on error.
func (q kvQuery) CountG(ctx context.Context) (int64, error) {
	return q.Count(ctx, boil.GetContextDB())
}

// Count returns the count of all KV records in the query.
func (q kvQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count kv rows")
	}

	return count, nil
}

// ExistsG checks if the row exists in the table, and panics on error.
func (q kvQuery) ExistsG(ctx context.Context) (bool, error) {
	return q.Exists(ctx, boil.GetContextDB())
}

// Exists checks if the row exists in the table.
func (q kvQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if kv exists")
	}

	return count > 0, nil
}

// KVS retrieves all the records using an executor.
func KVS(mods ...qm.QueryMod) kvQuery {
	mods = append(mods, qm.From("`kv`"))
	return kvQuery{NewQuery(mods...)}
}

// FindKVG retrieves a single record by ID.
func FindKVG(ctx context.Context, k string, selectCols ...string) (*KV, error) {
	return FindKV(ctx, boil.GetContextDB(), k, selectCols...)
}

// FindKV retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindKV(ctx context.Context, exec boil.ContextExecutor, k string, selectCols ...string) (*KV, error) {
	kvObj := &KV{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from `kv` where `k`=?", sel,
	)

	q := queries.Raw(query, k)

	err := q.Bind(ctx, exec, kvObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from kv")
	}

	return kvObj, nil
}

// InsertG a single record. See Insert for whitelist behavior description.
func (o *KV) InsertG(ctx context.Context, columns boil.Columns) error {
	return o.Insert(ctx, boil.GetContextDB(), columns)
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *KV) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no kv provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
		if o.UpdatedAt.IsZero() {
			o.UpdatedAt = currTime
		}
	}

	nzDefaults := queries.NonZeroDefaultSet(kvColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	kvInsertCacheMut.RLock()
	cache, cached := kvInsertCache[key]
	kvInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			kvAllColumns,
			kvColumnsWithDefault,
			kvColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(kvType, kvMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(kvType, kvMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO `kv` (`%s`) %%sVALUES (%s)%%s", strings.Join(wl, "`,`"), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO `kv` () VALUES ()%s%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			cache.retQuery = fmt.Sprintf("SELECT `%s` FROM `kv` WHERE %s", strings.Join(returnColumns, "`,`"), strmangle.WhereClause("`", "`", 0, kvPrimaryKeyColumns))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, vals)
	}

	_, err = exec.ExecContext(ctx, cache.query, vals...)

	if err != nil {
		return errors.Wrap(err, "boiler: unable to insert into kv")
	}

	var identifierCols []interface{}

	if len(cache.retMapping) == 0 {
		goto CacheNoHooks
	}

	identifierCols = []interface{}{
		o.K,
	}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.retQuery)
		fmt.Fprintln(boil.DebugWriter, identifierCols...)
	}

	err = exec.QueryRowContext(ctx, cache.retQuery, identifierCols...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to populate default values for kv")
	}

CacheNoHooks:
	if !cached {
		kvInsertCacheMut.Lock()
		kvInsertCache[key] = cache
		kvInsertCacheMut.Unlock()
	}

	return nil
}

// UpdateG a single KV record using the global executor.
// See Update for more documentation.
func (o *KV) UpdateG(ctx context.Context, columns boil.Columns) error {
	return o.Update(ctx, boil.GetContextDB(), columns)
}

// Update uses an executor to update the KV.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *KV) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
	}

	var err error
	key := makeCacheKey(columns, nil)
	kvUpdateCacheMut.RLock()
	cache, cached := kvUpdateCache[key]
	kvUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			kvAllColumns,
			kvPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return errors.New("boiler: unable to update kv, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE `kv` SET %s WHERE %s",
			strmangle.SetParamNames("`", "`", 0, wl),
			strmangle.WhereClause("`", "`", 0, kvPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(kvType, kvMapping, append(wl, kvPrimaryKeyColumns...))
		if err != nil {
			return err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, values)
	}

	_, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to update kv row")
	}

	if !cached {
		kvUpdateCacheMut.Lock()
		kvUpdateCache[key] = cache
		kvUpdateCacheMut.Unlock()
	}

	return nil
}

// UpdateAllG updates all rows with the specified column values.
func (q kvQuery) UpdateAllG(ctx context.Context, cols M) error {
	return q.UpdateAll(ctx, boil.GetContextDB(), cols)
}

// UpdateAll updates all rows with the specified column values.
func (q kvQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) error {
	queries.SetUpdate(q.Query, cols)

	_, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to update all for kv")
	}

	return nil
}

// UpdateAllG updates all rows with the specified column values.
func (o KVSlice) UpdateAllG(ctx context.Context, cols M) error {
	return o.UpdateAll(ctx, boil.GetContextDB(), cols)
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o KVSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) error {
	ln := int64(len(o))
	if ln == 0 {
		return nil
	}

	if len(cols) == 0 {
		return errors.New("boiler: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), kvPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE `kv` SET %s WHERE %s",
		strmangle.SetParamNames("`", "`", 0, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, kvPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}

	_, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to update all in kv slice")
	}

	return nil
}

// UpsertG attempts an insert, and does an update or ignore on conflict.
func (o *KV) UpsertG(ctx context.Context, updateColumns, insertColumns boil.Columns) error {
	return o.Upsert(ctx, boil.GetContextDB(), updateColumns, insertColumns)
}

var mySQLKVUniqueColumns = []string{
	"k",
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *KV) Upsert(ctx context.Context, exec boil.ContextExecutor, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no kv provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
		o.UpdatedAt = currTime
	}

	nzDefaults := queries.NonZeroDefaultSet(kvColumnsWithDefault, o)
	nzUniques := queries.NonZeroDefaultSet(mySQLKVUniqueColumns, o)

	if len(nzUniques) == 0 {
		return errors.New("cannot upsert with a table that cannot conflict on a unique column")
	}

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzUniques {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	kvUpsertCacheMut.RLock()
	cache, cached := kvUpsertCache[key]
	kvUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			kvAllColumns,
			kvColumnsWithDefault,
			kvColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			kvAllColumns,
			kvPrimaryKeyColumns,
		)

		if len(update) == 0 {
			return errors.New("boiler: unable to upsert kv, could not build update column list")
		}

		ret = strmangle.SetComplement(ret, nzUniques)
		cache.query = buildUpsertQueryMySQL(dialect, "kv", update, insert)
		cache.retQuery = fmt.Sprintf(
			"SELECT %s FROM `kv` WHERE %s",
			strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, ret), ","),
			strmangle.WhereClause("`", "`", 0, nzUniques),
		)

		cache.valueMapping, err = queries.BindMapping(kvType, kvMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(kvType, kvMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, vals)
	}

	_, err = exec.ExecContext(ctx, cache.query, vals...)

	if err != nil {
		return errors.Wrap(err, "boiler: unable to upsert for kv")
	}

	var uniqueMap []uint64
	var nzUniqueCols []interface{}

	if len(cache.retMapping) == 0 {
		goto CacheNoHooks
	}

	uniqueMap, err = queries.BindMapping(kvType, kvMapping, nzUniques)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to retrieve unique values for kv")
	}
	nzUniqueCols = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), uniqueMap)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.retQuery)
		fmt.Fprintln(boil.DebugWriter, nzUniqueCols...)
	}

	err = exec.QueryRowContext(ctx, cache.retQuery, nzUniqueCols...).Scan(returns...)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to populate default values for kv")
	}

CacheNoHooks:
	if !cached {
		kvUpsertCacheMut.Lock()
		kvUpsertCache[key] = cache
		kvUpsertCacheMut.Unlock()
	}

	return nil
}

// DeleteG deletes a single KV record.
// DeleteG will match against the primary key column to find the record to delete.
func (o *KV) DeleteG(ctx context.Context) error {
	return o.Delete(ctx, boil.GetContextDB())
}

// Delete deletes a single KV record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *KV) Delete(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("boiler: no KV provided for delete")
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), kvPrimaryKeyMapping)
	sql := "DELETE FROM `kv` WHERE `k`=?"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}

	_, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to delete from kv")
	}

	return nil
}

// DeleteAll deletes all matching rows.
func (q kvQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) error {
	if q.Query == nil {
		return errors.New("boiler: no kvQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	_, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to delete all from kv")
	}

	return nil
}

// DeleteAllG deletes all rows in the slice.
func (o KVSlice) DeleteAllG(ctx context.Context) error {
	return o.DeleteAll(ctx, boil.GetContextDB())
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o KVSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) error {
	if len(o) == 0 {
		return nil
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), kvPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM `kv` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, kvPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}

	_, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to delete all from kv slice")
	}

	return nil
}

// ReloadG refetches the object from the database using the primary keys.
func (o *KV) ReloadG(ctx context.Context) error {
	if o == nil {
		return errors.New("boiler: no KV provided for reload")
	}

	return o.Reload(ctx, boil.GetContextDB())
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *KV) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindKV(ctx, exec, o.K)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAllG refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *KVSlice) ReloadAllG(ctx context.Context) error {
	if o == nil {
		return errors.New("boiler: empty KVSlice provided for reload all")
	}

	return o.ReloadAll(ctx, boil.GetContextDB())
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *KVSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := KVSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), kvPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT `kv`.* FROM `kv` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, kvPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in KVSlice")
	}

	*o = slice

	return nil
}

// KVExistsG checks if the KV row exists.
func KVExistsG(ctx context.Context, k string) (bool, error) {
	return KVExists(ctx, boil.GetContextDB(), k)
}

// KVExists checks if the KV row exists.
func KVExists(ctx context.Context, exec boil.ContextExecutor, k string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from `kv` where `k`=? limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, k)
	}

	row := exec.QueryRowContext(ctx, sql, k)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if kv exists")
	}

	return exists, nil
}
