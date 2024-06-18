package core

import (
	"context"
	"fmt"
	"reflect"
	"xo-packs/db"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

// StructToUpdateValues converts a struct into a map of column names and their values.
func StructToUpdateValues(data interface{}) map[string]interface{} {
	values := make(map[string]interface{})
	valueType := reflect.ValueOf(data)
	valueType = reflect.Indirect(valueType)

	for i := 0; i < valueType.NumField(); i++ {
		field := valueType.Type().Field(i)
		columnName := field.Tag.Get("db")
		if columnName != "" {
			fieldVal := valueType.Field(i)
			if fieldVal.Kind() == reflect.Ptr {
				if fieldVal.IsNil() {
					values[columnName] = fieldVal.Interface()
				} else {
					values[columnName] = fieldVal.Elem().Interface()
				}
			} else {
				values[columnName] = valueType.Field(i).Interface()
			}
		}
	}
	return values
}

func ModelColumns(data interface{}) []string {
	valueType := reflect.ValueOf(data)
	valueType = reflect.Indirect(valueType)
	numFields := valueType.NumField()
	var cols []string

	for i := 0; i < numFields; i++ {
		fieldVal := valueType.Field(i)
		field := valueType.Type().Field(i)
		if !fieldVal.IsNil() {
			columnName := field.Tag.Get("db")
			cols = append(cols, columnName)
		}
	}

	return cols
}

func StructValues(data interface{}) []interface{} {
	valueType := reflect.ValueOf(data)
	valueType = reflect.Indirect(valueType)
	numFields := valueType.NumField()
	var values []interface{}

	for i := 0; i < numFields; i++ {
		fieldVal := valueType.Field(i)
		if !fieldVal.IsNil() {
			values = append(values, valueType.Field(i).Interface())
		}
	}

	return values
}

func GetSortMappings(ctx context.Context, category *string, tx *sqlx.Tx, err error) (map[string]string, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Select("sort_alias", "sort_field").
		From(db.SCHEMA_SORT_MAPPINGS).
		Where(squirrel.Eq{"sort_category": *category}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	alias := ""
	field := ""
	mapping := map[string]string{}
	for rows.Next() {
		if err = rows.Scan(&alias, &field); err != nil {
			return nil, err
		}
		mapping[alias] = field
	}
	if len(mapping) <= 0 {
		return nil, &ErrorResp{Message: fmt.Sprintf("Critical error: mapping for sort category '%v' is empty", *category)}
	}
	return mapping, nil
}
