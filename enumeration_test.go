//go:build experimental

package tiledb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnumeration(t *testing.T) {
	config, err := NewConfig()
	require.NoError(t, err)
	tdbCtx, err := NewContext(config)
	require.NoError(t, err)

	romanNumerals, err := NewOrderedEnumeration(tdbCtx, "romanNumerals",
		[]string{"i", "ii", "iii", "iv", "v", "vi", "vii", "viii", "ix", "x", "xi", "xii", "xiii", "xiv", "xv", "xvi"})
	require.NoError(t, err)

	powersOfTwo, err := NewOrderedEnumeration(tdbCtx, "powersOfTwo", []uint32{1, 2, 4, 8, 16, 32, 64, 128, 256})
	require.NoError(t, err)

	truth, err := NewUnorderedEnumeration(tdbCtx, "truth", []bool{false, true})
	require.NoError(t, err)

	t.Run("Name", func(t *testing.T) {
		romanName, err := romanNumerals.Name()
		require.NoError(t, err)
		assert.Equal(t, "romanNumerals", romanName)

		powersName, err := powersOfTwo.Name()
		require.NoError(t, err)
		assert.Equal(t, "powersOfTwo", powersName)
	})

	t.Run("Type", func(t *testing.T) {
		romanType, err := romanNumerals.Type()
		require.NoError(t, err)
		assert.Equal(t, TILEDB_STRING_ASCII, romanType)

		powersType, err := powersOfTwo.Type()
		require.NoError(t, err)
		assert.Equal(t, TILEDB_UINT32, powersType)
	})

	t.Run("CellValNum", func(t *testing.T) {
		romanCellNum, err := romanNumerals.CellValNum()
		require.NoError(t, err)
		assert.Equal(t, TILEDB_VAR_NUM, romanCellNum)

		powersCellNum, err := powersOfTwo.CellValNum()
		require.NoError(t, err)
		assert.Equal(t, uint32(1), powersCellNum)
	})

	t.Run("IsOrdered", func(t *testing.T) {
		romanOrdered, err := romanNumerals.IsOrdered()
		require.NoError(t, err)
		assert.Equal(t, true, romanOrdered)

		powersOrdered, err := powersOfTwo.IsOrdered()
		require.NoError(t, err)
		assert.Equal(t, true, powersOrdered)

		truthOrdered, err := truth.IsOrdered()
		require.NoError(t, err)
		assert.Equal(t, false, truthOrdered)
	})

	t.Run("Values", func(t *testing.T) {
		romanValues, err := romanNumerals.Values()
		require.NoError(t, err)
		romanStrings := romanValues.([]string)
		assert.Equal(t, romanStrings, []string{"i", "ii", "iii", "iv", "v", "vi", "vii", "viii", "ix", "x", "xi", "xii", "xiii", "xiv", "xv", "xvi"})

		powerValues, err := powersOfTwo.Values()
		require.NoError(t, err)
		powersInts := powerValues.([]uint32)
		assert.Equal(t, powersInts, []uint32{1, 2, 4, 8, 16, 32, 64, 128, 256})

		truthValues, err := truth.Values()
		require.NoError(t, err)
		truthBools := truthValues.([]bool)
		assert.Equal(t, truthBools, []bool{false, true})
	})

	t.Run("Dump", func(t *testing.T) {
		fname := filepath.Join(t.TempDir(), "roman-dump")
		require.NoError(t, romanNumerals.Dump(fname))

		dump, err := os.ReadFile(fname)
		require.NoError(t, err)
		contents := string(dump)
		assert.Contains(t, contents, "Name: romanNumerals")
		assert.Contains(t, contents, "Element Count: 16")
	})
}

func TestEnumerationAndSchema(t *testing.T) {
	schema := arraySchemaWithEnumerations(t)

	config, err := NewConfig()
	require.NoError(t, err)
	tdbCtx, err := NewContext(config)
	require.NoError(t, err)

	arrayPath := t.TempDir()
	array, err := NewArray(tdbCtx, arrayPath)
	require.NoError(t, err)
	require.NoError(t, array.Create(schema))
	require.NoError(t, array.Open(TILEDB_READ))
	t.Cleanup(func() { array.Close() })

	t.Run("FromArray", func(t *testing.T) {
		romanEnum, err := array.GetEnumeration("romanNumerals")
		require.NoError(t, err)
		romanName, err := romanEnum.Name()
		require.Equal(t, "romanNumerals", romanName)
	})

	t.Run("FromAttribute", func(t *testing.T) {
		romanAttr, err := schema.AttributeFromName("roman")
		require.NoError(t, err)
		romanName, err := romanAttr.GetEnumerationName()
		require.NoError(t, err)
		assert.Equal(t, "romanNumerals", romanName)
	})
}

func TestEnumerationQueryCondition(t *testing.T) {
	schema := arraySchemaWithEnumerations(t)

	config, err := NewConfig()
	require.NoError(t, err)
	tdbCtx, err := NewContext(config)
	require.NoError(t, err)

	arrayPath := t.TempDir()
	array, err := NewArray(tdbCtx, arrayPath)
	require.NoError(t, err)
	require.NoError(t, array.Create(schema))

	//=====
	// write to the array. Each cell gets the row order rank.
	// The array will look like
	//  0  1  2  3
	//  4  5  6  7
	//  8  9 10 11
	// 12 13 14 15

	array, err = NewArray(tdbCtx, arrayPath)
	require.NoError(t, err)
	require.NoError(t, array.Open(TILEDB_WRITE))
	wQuery, err := NewQuery(tdbCtx, array)
	require.NoError(t, err)
	_, err = wQuery.SetDataBuffer("rows", []uint8{1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4})
	require.NoError(t, err)
	_, err = wQuery.SetDataBuffer("cols", []uint8{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4})
	require.NoError(t, err)
	_, err = wQuery.SetDataBuffer("greek", []uint8{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15})
	require.NoError(t, err)
	_, err = wQuery.SetDataBuffer("roman", []uint8{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15})
	require.NoError(t, err)
	require.NoError(t, wQuery.Submit())
	require.NoError(t, array.Close())

	t.Run("EnumerationsEnabled", func(t *testing.T) {
		array, err := NewArray(tdbCtx, arrayPath)
		require.NoError(t, err)
		require.NoError(t, array.Open(TILEDB_READ))
		rQuery, err := NewQuery(tdbCtx, array)
		require.NoError(t, err)
		qcR, err := NewQueryCondition(tdbCtx, "roman", TILEDB_QUERY_CONDITION_EQ, "vi")
		require.NoError(t, err)
		qcG, err := NewQueryCondition(tdbCtx, "greek", TILEDB_QUERY_CONDITION_EQ, "β")
		require.NoError(t, err)
		qc, err := NewQueryConditionCombination(tdbCtx, qcR, TILEDB_QUERY_CONDITION_OR, qcG)
		require.NoError(t, err)
		require.NoError(t, rQuery.SetQueryCondition(qc))

		rowsBuffer := make([]uint8, 16)
		_, err = rQuery.SetDataBuffer("rows", rowsBuffer)
		require.NoError(t, err)
		colsBuffer := make([]uint8, 16)
		_, err = rQuery.SetDataBuffer("cols", colsBuffer)
		require.NoError(t, err)
		greekBuffer := make([]uint8, 16)
		_, err = rQuery.SetDataBuffer("greek", greekBuffer)
		require.NoError(t, err)
		romanBuffer := make([]uint8, 16)
		_, err = rQuery.SetDataBuffer("roman", romanBuffer)
		require.NoError(t, err)

		require.NoError(t, rQuery.Submit())
		require.NoError(t, array.Close())

		assert.Equal(t, []uint8{1, 2, 0}, rowsBuffer[0:3])
		assert.Equal(t, []uint8{2, 2, 0}, colsBuffer[0:3])
		assert.Equal(t, []uint8{1, 5, 0}, greekBuffer[0:3])
		assert.Equal(t, []uint8{1, 5, 0}, romanBuffer[0:3])
	})

	t.Run("EnumerationsDisabled", func(t *testing.T) {
		// note that when query conditions don't use enumerations, we must
		// shift the query values by 1 to get the same result as with enumerations enabled
		// This is because enumerations indices start at 0

		array, err := NewArray(tdbCtx, arrayPath)
		require.NoError(t, err)
		require.NoError(t, array.Open(TILEDB_READ))
		rQuery, err := NewQuery(tdbCtx, array)
		require.NoError(t, err)
		qcR, err := NewQueryCondition(tdbCtx, "roman", TILEDB_QUERY_CONDITION_EQ, uint8(5)) // 5 instead of vi(6)
		require.NoError(t, err)
		require.NoError(t, qcR.UseEnumeration(false))
		qcG, err := NewQueryCondition(tdbCtx, "greek", TILEDB_QUERY_CONDITION_EQ, uint8(1)) // 1 instead of β(2)
		require.NoError(t, err)
		require.NoError(t, qcG.UseEnumeration(false))
		qc, err := NewQueryConditionCombination(tdbCtx, qcR, TILEDB_QUERY_CONDITION_OR, qcG)
		require.NoError(t, err)
		require.NoError(t, qc.UseEnumeration(false))
		require.NoError(t, rQuery.SetQueryCondition(qc))

		rowsBuffer := make([]uint8, 16)
		_, err = rQuery.SetDataBuffer("rows", rowsBuffer)
		require.NoError(t, err)
		colsBuffer := make([]uint8, 16)
		_, err = rQuery.SetDataBuffer("cols", colsBuffer)
		require.NoError(t, err)
		greekBuffer := make([]uint8, 16)
		_, err = rQuery.SetDataBuffer("greek", greekBuffer)
		require.NoError(t, err)
		romanBuffer := make([]uint8, 16)
		_, err = rQuery.SetDataBuffer("roman", romanBuffer)
		require.NoError(t, err)

		require.NoError(t, rQuery.Submit())
		require.NoError(t, array.Close())

		assert.Equal(t, []uint8{1, 2, 0}, rowsBuffer[0:3])
		assert.Equal(t, []uint8{2, 2, 0}, colsBuffer[0:3])
		assert.Equal(t, []uint8{1, 5, 0}, greekBuffer[0:3])
		assert.Equal(t, []uint8{1, 5, 0}, romanBuffer[0:3])
	})

	t.Run("LabelNotExists", func(t *testing.T) {
		array, err := NewArray(tdbCtx, arrayPath)
		require.NoError(t, err)
		require.NoError(t, array.Open(TILEDB_READ))
		rQuery, err := NewQuery(tdbCtx, array)
		require.NoError(t, err)
		qcR, err := NewQueryCondition(tdbCtx, "roman", TILEDB_QUERY_CONDITION_EQ, "C")
		require.NoError(t, err)
		require.NoError(t, rQuery.SetQueryCondition(qcR))

		rowsBuffer := make([]uint8, 16)
		_, err = rQuery.SetDataBuffer("rows", rowsBuffer)
		require.NoError(t, err)
		colsBuffer := make([]uint8, 16)
		_, err = rQuery.SetDataBuffer("cols", colsBuffer)
		require.NoError(t, err)
		err = rQuery.Submit()
		require.Error(t, err)
		require.Contains(t, err.Error(), "Enumeration value not found")
		require.NoError(t, array.Close())
	})
}

func TestEnumerationEmpty(t *testing.T) {
	schema := arraySchemaWithEmptyEnumerations(t)

	config, err := NewConfig()
	require.NoError(t, err)
	tdbCtx, err := NewContext(config)
	require.NoError(t, err)

	arrayPath := t.TempDir()
	array, err := NewArray(tdbCtx, arrayPath)
	require.NoError(t, err)
	require.NoError(t, array.Create(schema))
}

func arraySchemaWithEnumerations(t *testing.T) *ArraySchema {
	config, err := NewConfig()
	require.NoError(t, err)
	tdbCtx, err := NewContext(config)
	require.NoError(t, err)

	//=====
	// create a sparse array [1,4]x[1,4]
	//

	schema, err := NewArraySchema(tdbCtx, TILEDB_SPARSE)
	require.NoError(t, err)
	require.NoError(t, schema.SetCellOrder(TILEDB_ROW_MAJOR))
	require.NoError(t, schema.SetTileOrder(TILEDB_ROW_MAJOR))

	domain, err := NewDomain(tdbCtx)
	require.NoError(t, err)
	dimRows, err := NewDimension(tdbCtx, "rows", TILEDB_UINT8, []uint8{1, 4}, uint8(2))
	require.NoError(t, err)
	dimCols, err := NewDimension(tdbCtx, "cols", TILEDB_UINT8, []uint8{1, 4}, uint8(2))
	require.NoError(t, err)
	require.NoError(t, domain.AddDimensions(dimRows, dimCols))
	require.NoError(t, schema.SetDomain(domain))

	greekNumerals, err := NewOrderedEnumeration(tdbCtx, "greekNumerals",
		[]string{"α", "β", "γ", "δ", "ε", "στ", "ζ", "η", "θ", "ι", "ια", "ιβ", "ιγ", "ιδ", "ιε", "ιστ"})
	require.NoError(t, err)
	require.NoError(t, schema.AddEnumeration(greekNumerals))
	romanNumerals, err := NewOrderedEnumeration(tdbCtx, "romanNumerals",
		[]string{"i", "ii", "iii", "iv", "v", "vi", "vii", "viii", "ix", "x", "xi", "xii", "xiii", "xiv", "xv", "xvi"})
	require.NoError(t, err)
	require.NoError(t, schema.AddEnumeration(romanNumerals))

	greekAttr, err := NewAttribute(tdbCtx, "greek", TILEDB_UINT8)
	require.NoError(t, err)
	require.NoError(t, greekAttr.SetEnumerationName("greekNumerals"))
	romanAttr, err := NewAttribute(tdbCtx, "roman", TILEDB_UINT8)
	require.NoError(t, err)
	require.NoError(t, romanAttr.SetEnumerationName("romanNumerals"))
	require.NoError(t, schema.AddAttributes(greekAttr, romanAttr))

	return schema
}

func arraySchemaWithEmptyEnumerations(t *testing.T) *ArraySchema {
	config, err := NewConfig()
	require.NoError(t, err)
	tdbCtx, err := NewContext(config)
	require.NoError(t, err)

	//=====
	// create a sparse array [1,4]x[1,4]
	//

	schema, err := NewArraySchema(tdbCtx, TILEDB_SPARSE)
	require.NoError(t, err)
	require.NoError(t, schema.SetCellOrder(TILEDB_ROW_MAJOR))
	require.NoError(t, schema.SetTileOrder(TILEDB_ROW_MAJOR))

	domain, err := NewDomain(tdbCtx)
	require.NoError(t, err)
	dimRows, err := NewDimension(tdbCtx, "rows", TILEDB_UINT8, []uint8{1, 4}, uint8(2))
	require.NoError(t, err)
	dimCols, err := NewDimension(tdbCtx, "cols", TILEDB_UINT8, []uint8{1, 4}, uint8(2))
	require.NoError(t, err)
	require.NoError(t, domain.AddDimensions(dimRows, dimCols))
	require.NoError(t, schema.SetDomain(domain))

	greekNumerals, err := NewOrderedEnumeration[string](tdbCtx, "greekNumerals", nil)
	require.NoError(t, err)
	require.NoError(t, schema.AddEnumeration(greekNumerals))
	romanNumerals, err := NewOrderedEnumeration[string](tdbCtx, "romanNumerals", nil)
	require.NoError(t, err)
	require.NoError(t, schema.AddEnumeration(romanNumerals))

	greekAttr, err := NewAttribute(tdbCtx, "greek", TILEDB_UINT8)
	require.NoError(t, err)
	require.NoError(t, greekAttr.SetEnumerationName("greekNumerals"))
	romanAttr, err := NewAttribute(tdbCtx, "roman", TILEDB_UINT8)
	require.NoError(t, err)
	require.NoError(t, romanAttr.SetEnumerationName("romanNumerals"))
	require.NoError(t, schema.AddAttributes(greekAttr, romanAttr))

	return schema
}
