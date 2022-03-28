package tiledb

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObjectArray(t *testing.T) {
	// Create context
	context, err := NewContext(nil)
	require.NoError(t, err)

	// create temp group name
	groupPath := t.TempDir()

	group, err := NewGroup(context, groupPath)
	require.NoError(t, err)

	// Create initial group
	require.NoError(t, group.Create())

	arrayGroupPath := filepath.Join(groupPath, "arrays")
	arrayGroup, err := NewGroup(context, groupPath)
	require.NoError(t, err)

	// Create the array group
	require.NoError(t, arrayGroup.Create())

	tmpArrayPath := filepath.Join(arrayGroupPath, "tiledb_test_array")

	// Create new array struct
	array, err := NewArray(context, tmpArrayPath)
	require.NoError(t, err)
	assert.NotNil(t, array)

	arraySchema := buildArraySchema(context, t)

	// Create array on disk
	require.NoError(t, array.Create(arraySchema))

	objType, err := ObjectType(context, groupPath)
	require.NoError(t, err)
	assert.Equal(t, TILEDB_GROUP, objType)

	objType, err = ObjectType(context, tmpArrayPath)
	require.NoError(t, err)
	assert.Equal(t, TILEDB_ARRAY, objType)

	objectList, err := ObjectWalk(context, groupPath, TILEDB_PREORDER)
	require.NoError(t, err)
	assert.Equal(t, 2, len(objectList.objectList))
	assert.Equal(t, TILEDB_GROUP, objectList.objectList[0].objectTypeEnum)
	assert.Equal(t, TILEDB_ARRAY, objectList.objectList[1].objectTypeEnum)

	objectList, err = ObjectLs(context, groupPath)
	require.NoError(t, err)
	assert.Equal(t, 1, len(objectList.objectList))
	assert.Equal(t, TILEDB_GROUP, objectList.objectList[0].objectTypeEnum)
}
