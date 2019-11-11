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
)

// CorporationDelta is an object representing the database table.
type CorporationDelta struct {
	ID            uint64    `db:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	CorporationID uint      `db:"corporation_id" boil:"corporation_id" json:"corporation_id" toml:"corporation_id" yaml:"corporation_id"`
	MemberCount   uint64    `db:"member_count" boil:"member_count" json:"member_count" toml:"member_count" yaml:"member_count"`
	CreatedAt     time.Time `db:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *corporationDeltaR `db:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L corporationDeltaL  `db:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var CorporationDeltaColumns = struct {
	ID            string
	CorporationID string
	MemberCount   string
	CreatedAt     string
}{
	ID:            "id",
	CorporationID: "corporation_id",
	MemberCount:   "member_count",
	CreatedAt:     "created_at",
}

// Generated where

var CorporationDeltaWhere = struct {
	ID            whereHelperuint64
	CorporationID whereHelperuint
	MemberCount   whereHelperuint64
	CreatedAt     whereHelpertime_Time
}{
	ID:            whereHelperuint64{field: "`corporation_deltas`.`id`"},
	CorporationID: whereHelperuint{field: "`corporation_deltas`.`corporation_id`"},
	MemberCount:   whereHelperuint64{field: "`corporation_deltas`.`member_count`"},
	CreatedAt:     whereHelpertime_Time{field: "`corporation_deltas`.`created_at`"},
}

// CorporationDeltaRels is where relationship names are stored.
var CorporationDeltaRels = struct {
}{}

// corporationDeltaR is where relationships are stored.
type corporationDeltaR struct {
}

// NewStruct creates a new relationship struct
func (*corporationDeltaR) NewStruct() *corporationDeltaR {
	return &corporationDeltaR{}
}

// corporationDeltaL is where Load methods for each relationship are stored.
type corporationDeltaL struct{}

var (
	corporationDeltaAllColumns            = []string{"id", "corporation_id", "member_count", "created_at"}
	corporationDeltaColumnsWithoutDefault = []string{"corporation_id", "member_count", "created_at"}
	corporationDeltaColumnsWithDefault    = []string{"id"}
	corporationDeltaPrimaryKeyColumns     = []string{"id"}
)

type (
	// CorporationDeltaSlice is an alias for a slice of pointers to CorporationDelta.
	// This should generally be used opposed to []CorporationDelta.
	CorporationDeltaSlice []*CorporationDelta

	corporationDeltaQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	corporationDeltaType                 = reflect.TypeOf(&CorporationDelta{})
	corporationDeltaMapping              = queries.MakeStructMapping(corporationDeltaType)
	corporationDeltaPrimaryKeyMapping, _ = queries.BindMapping(corporationDeltaType, corporationDeltaMapping, corporationDeltaPrimaryKeyColumns)
	corporationDeltaInsertCacheMut       sync.RWMutex
	corporationDeltaInsertCache          = make(map[string]insertCache)
	corporationDeltaUpdateCacheMut       sync.RWMutex
	corporationDeltaUpdateCache          = make(map[string]updateCache)
	corporationDeltaUpsertCacheMut       sync.RWMutex
	corporationDeltaUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

// OneG returns a single corporationDelta record from the query using the global executor.
func (q corporationDeltaQuery) OneG(ctx context.Context) (*CorporationDelta, error) {
	return q.One(ctx, boil.GetContextDB())
}

// One returns a single corporationDelta record from the query.
func (q corporationDeltaQuery) One(ctx context.Context, exec boil.ContextExecutor) (*CorporationDelta, error) {
	o := &CorporationDelta{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for corporation_deltas")
	}

	return o, nil
}

// AllG returns all CorporationDelta records from the query using the global executor.
func (q corporationDeltaQuery) AllG(ctx context.Context) (CorporationDeltaSlice, error) {
	return q.All(ctx, boil.GetContextDB())
}

// All returns all CorporationDelta records from the query.
func (q corporationDeltaQuery) All(ctx context.Context, exec boil.ContextExecutor) (CorporationDeltaSlice, error) {
	var o []*CorporationDelta

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to CorporationDelta slice")
	}

	return o, nil
}

// CountG returns the count of all CorporationDelta records in the query, and panics on error.
func (q corporationDeltaQuery) CountG(ctx context.Context) (int64, error) {
	return q.Count(ctx, boil.GetContextDB())
}

// Count returns the count of all CorporationDelta records in the query.
func (q corporationDeltaQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count corporation_deltas rows")
	}

	return count, nil
}

// ExistsG checks if the row exists in the table, and panics on error.
func (q corporationDeltaQuery) ExistsG(ctx context.Context) (bool, error) {
	return q.Exists(ctx, boil.GetContextDB())
}

// Exists checks if the row exists in the table.
func (q corporationDeltaQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if corporation_deltas exists")
	}

	return count > 0, nil
}

// CorporationDeltas retrieves all the records using an executor.
func CorporationDeltas(mods ...qm.QueryMod) corporationDeltaQuery {
	mods = append(mods, qm.From("`corporation_deltas`"))
	return corporationDeltaQuery{NewQuery(mods...)}
}

// FindCorporationDeltaG retrieves a single record by ID.
func FindCorporationDeltaG(ctx context.Context, iD uint64, selectCols ...string) (*CorporationDelta, error) {
	return FindCorporationDelta(ctx, boil.GetContextDB(), iD, selectCols...)
}

// FindCorporationDelta retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindCorporationDelta(ctx context.Context, exec boil.ContextExecutor, iD uint64, selectCols ...string) (*CorporationDelta, error) {
	corporationDeltaObj := &CorporationDelta{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from `corporation_deltas` where `id`=?", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, corporationDeltaObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from corporation_deltas")
	}

	return corporationDeltaObj, nil
}

// InsertG a single record. See Insert for whitelist behavior description.
func (o *CorporationDelta) InsertG(ctx context.Context, columns boil.Columns) error {
	return o.Insert(ctx, boil.GetContextDB(), columns)
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *CorporationDelta) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no corporation_deltas provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	nzDefaults := queries.NonZeroDefaultSet(corporationDeltaColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	corporationDeltaInsertCacheMut.RLock()
	cache, cached := corporationDeltaInsertCache[key]
	corporationDeltaInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			corporationDeltaAllColumns,
			corporationDeltaColumnsWithDefault,
			corporationDeltaColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(corporationDeltaType, corporationDeltaMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(corporationDeltaType, corporationDeltaMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO `corporation_deltas` (`%s`) %%sVALUES (%s)%%s", strings.Join(wl, "`,`"), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO `corporation_deltas` () VALUES ()%s%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			cache.retQuery = fmt.Sprintf("SELECT `%s` FROM `corporation_deltas` WHERE %s", strings.Join(returnColumns, "`,`"), strmangle.WhereClause("`", "`", 0, corporationDeltaPrimaryKeyColumns))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, vals)
	}

	result, err := exec.ExecContext(ctx, cache.query, vals...)

	if err != nil {
		return errors.Wrap(err, "boiler: unable to insert into corporation_deltas")
	}

	var lastID int64
	var identifierCols []interface{}

	if len(cache.retMapping) == 0 {
		goto CacheNoHooks
	}

	lastID, err = result.LastInsertId()
	if err != nil {
		return ErrSyncFail
	}

	o.ID = uint64(lastID)
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == corporationDeltaMapping["ID"] {
		goto CacheNoHooks
	}

	identifierCols = []interface{}{
		o.ID,
	}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.retQuery)
		fmt.Fprintln(boil.DebugWriter, identifierCols...)
	}

	err = exec.QueryRowContext(ctx, cache.retQuery, identifierCols...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to populate default values for corporation_deltas")
	}

CacheNoHooks:
	if !cached {
		corporationDeltaInsertCacheMut.Lock()
		corporationDeltaInsertCache[key] = cache
		corporationDeltaInsertCacheMut.Unlock()
	}

	return nil
}

// UpdateG a single CorporationDelta record using the global executor.
// See Update for more documentation.
func (o *CorporationDelta) UpdateG(ctx context.Context, columns boil.Columns) error {
	return o.Update(ctx, boil.GetContextDB(), columns)
}

// Update uses an executor to update the CorporationDelta.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *CorporationDelta) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	var err error
	key := makeCacheKey(columns, nil)
	corporationDeltaUpdateCacheMut.RLock()
	cache, cached := corporationDeltaUpdateCache[key]
	corporationDeltaUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			corporationDeltaAllColumns,
			corporationDeltaPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return errors.New("boiler: unable to update corporation_deltas, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE `corporation_deltas` SET %s WHERE %s",
			strmangle.SetParamNames("`", "`", 0, wl),
			strmangle.WhereClause("`", "`", 0, corporationDeltaPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(corporationDeltaType, corporationDeltaMapping, append(wl, corporationDeltaPrimaryKeyColumns...))
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
		return errors.Wrap(err, "boiler: unable to update corporation_deltas row")
	}

	if !cached {
		corporationDeltaUpdateCacheMut.Lock()
		corporationDeltaUpdateCache[key] = cache
		corporationDeltaUpdateCacheMut.Unlock()
	}

	return nil
}

// UpdateAllG updates all rows with the specified column values.
func (q corporationDeltaQuery) UpdateAllG(ctx context.Context, cols M) error {
	return q.UpdateAll(ctx, boil.GetContextDB(), cols)
}

// UpdateAll updates all rows with the specified column values.
func (q corporationDeltaQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) error {
	queries.SetUpdate(q.Query, cols)

	_, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to update all for corporation_deltas")
	}

	return nil
}

// UpdateAllG updates all rows with the specified column values.
func (o CorporationDeltaSlice) UpdateAllG(ctx context.Context, cols M) error {
	return o.UpdateAll(ctx, boil.GetContextDB(), cols)
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o CorporationDeltaSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) error {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), corporationDeltaPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE `corporation_deltas` SET %s WHERE %s",
		strmangle.SetParamNames("`", "`", 0, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, corporationDeltaPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}

	_, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to update all in corporationDelta slice")
	}

	return nil
}

// UpsertG attempts an insert, and does an update or ignore on conflict.
func (o *CorporationDelta) UpsertG(ctx context.Context, updateColumns, insertColumns boil.Columns) error {
	return o.Upsert(ctx, boil.GetContextDB(), updateColumns, insertColumns)
}

var mySQLCorporationDeltaUniqueColumns = []string{
	"id",
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *CorporationDelta) Upsert(ctx context.Context, exec boil.ContextExecutor, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no corporation_deltas provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	nzDefaults := queries.NonZeroDefaultSet(corporationDeltaColumnsWithDefault, o)
	nzUniques := queries.NonZeroDefaultSet(mySQLCorporationDeltaUniqueColumns, o)

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

	corporationDeltaUpsertCacheMut.RLock()
	cache, cached := corporationDeltaUpsertCache[key]
	corporationDeltaUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			corporationDeltaAllColumns,
			corporationDeltaColumnsWithDefault,
			corporationDeltaColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			corporationDeltaAllColumns,
			corporationDeltaPrimaryKeyColumns,
		)

		if len(update) == 0 {
			return errors.New("boiler: unable to upsert corporation_deltas, could not build update column list")
		}

		ret = strmangle.SetComplement(ret, nzUniques)
		cache.query = buildUpsertQueryMySQL(dialect, "corporation_deltas", update, insert)
		cache.retQuery = fmt.Sprintf(
			"SELECT %s FROM `corporation_deltas` WHERE %s",
			strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, ret), ","),
			strmangle.WhereClause("`", "`", 0, nzUniques),
		)

		cache.valueMapping, err = queries.BindMapping(corporationDeltaType, corporationDeltaMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(corporationDeltaType, corporationDeltaMapping, ret)
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

	result, err := exec.ExecContext(ctx, cache.query, vals...)

	if err != nil {
		return errors.Wrap(err, "boiler: unable to upsert for corporation_deltas")
	}

	var lastID int64
	var uniqueMap []uint64
	var nzUniqueCols []interface{}

	if len(cache.retMapping) == 0 {
		goto CacheNoHooks
	}

	lastID, err = result.LastInsertId()
	if err != nil {
		return ErrSyncFail
	}

	o.ID = uint64(lastID)
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == corporationDeltaMapping["id"] {
		goto CacheNoHooks
	}

	uniqueMap, err = queries.BindMapping(corporationDeltaType, corporationDeltaMapping, nzUniques)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to retrieve unique values for corporation_deltas")
	}
	nzUniqueCols = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), uniqueMap)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.retQuery)
		fmt.Fprintln(boil.DebugWriter, nzUniqueCols...)
	}

	err = exec.QueryRowContext(ctx, cache.retQuery, nzUniqueCols...).Scan(returns...)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to populate default values for corporation_deltas")
	}

CacheNoHooks:
	if !cached {
		corporationDeltaUpsertCacheMut.Lock()
		corporationDeltaUpsertCache[key] = cache
		corporationDeltaUpsertCacheMut.Unlock()
	}

	return nil
}

// DeleteG deletes a single CorporationDelta record.
// DeleteG will match against the primary key column to find the record to delete.
func (o *CorporationDelta) DeleteG(ctx context.Context) error {
	return o.Delete(ctx, boil.GetContextDB())
}

// Delete deletes a single CorporationDelta record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *CorporationDelta) Delete(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("boiler: no CorporationDelta provided for delete")
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), corporationDeltaPrimaryKeyMapping)
	sql := "DELETE FROM `corporation_deltas` WHERE `id`=?"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}

	_, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to delete from corporation_deltas")
	}

	return nil
}

// DeleteAll deletes all matching rows.
func (q corporationDeltaQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) error {
	if q.Query == nil {
		return errors.New("boiler: no corporationDeltaQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	_, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to delete all from corporation_deltas")
	}

	return nil
}

// DeleteAllG deletes all rows in the slice.
func (o CorporationDeltaSlice) DeleteAllG(ctx context.Context) error {
	return o.DeleteAll(ctx, boil.GetContextDB())
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o CorporationDeltaSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) error {
	if len(o) == 0 {
		return nil
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), corporationDeltaPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM `corporation_deltas` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, corporationDeltaPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}

	_, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to delete all from corporationDelta slice")
	}

	return nil
}

// ReloadG refetches the object from the database using the primary keys.
func (o *CorporationDelta) ReloadG(ctx context.Context) error {
	if o == nil {
		return errors.New("boiler: no CorporationDelta provided for reload")
	}

	return o.Reload(ctx, boil.GetContextDB())
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *CorporationDelta) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindCorporationDelta(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAllG refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *CorporationDeltaSlice) ReloadAllG(ctx context.Context) error {
	if o == nil {
		return errors.New("boiler: empty CorporationDeltaSlice provided for reload all")
	}

	return o.ReloadAll(ctx, boil.GetContextDB())
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *CorporationDeltaSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := CorporationDeltaSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), corporationDeltaPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT `corporation_deltas`.* FROM `corporation_deltas` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, corporationDeltaPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in CorporationDeltaSlice")
	}

	*o = slice

	return nil
}

// CorporationDeltaExistsG checks if the CorporationDelta row exists.
func CorporationDeltaExistsG(ctx context.Context, iD uint64) (bool, error) {
	return CorporationDeltaExists(ctx, boil.GetContextDB(), iD)
}

// CorporationDeltaExists checks if the CorporationDelta row exists.
func CorporationDeltaExists(ctx context.Context, exec boil.ContextExecutor, iD uint64) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from `corporation_deltas` where `id`=? limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}

	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if corporation_deltas exists")
	}

	return exists, nil
}
