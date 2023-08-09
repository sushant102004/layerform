package layers

import (
	"context"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ergomake/layerform/internal/data/model"
)

func setup(t *testing.T, layers []*model.Layer) *filebackend {
	ctx := context.Background()
	tmpDir := t.TempDir()
	fpath := path.Join(tmpDir, "layerform.definitions.json")
	backend, err := NewFileBackend(ctx, fpath)
	require.NoError(t, err)

	err = backend.UpdateLayers(ctx, layers)
	require.NoError(t, err)
	return backend

}

func TestLayers_FileBackend(t *testing.T) {
	layers := []*model.Layer{
		{Name: "layer1"},
		{Name: "layer2"},
	}
	stateBackend := setup(t, layers)

	layer1, err := stateBackend.GetLayer(context.Background(), "layer1")
	require.NoError(t, err)
	assert.Equal(t, layers[0], layer1)

	layer2, err := stateBackend.GetLayer(context.Background(), "layer2")
	require.NoError(t, err)
	assert.Equal(t, layers[1], layer2)

	_, err = stateBackend.GetLayer(context.Background(), "layer3")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestFileBackend_ResolveDependencies(t *testing.T) {
	layer1 := &model.Layer{Name: "layer1", Dependencies: []string{"layer2"}}
	layer2 := &model.Layer{Name: "layer2", Dependencies: []string{"layer3"}}
	layer3 := &model.Layer{Name: "layer3", Dependencies: []string{"layer4"}}

	stateBackend := setup(t, []*model.Layer{layer1, layer2, layer3})

	t.Run("single dependency", func(t *testing.T) {
		dependencies, err := stateBackend.ResolveDependencies(context.Background(), layer1)
		require.NoError(t, err)
		assert.Len(t, dependencies, 1)
		assert.Equal(t, layer2, dependencies[0])
	})

	t.Run("dependency not found", func(t *testing.T) {
		_, err := stateBackend.ResolveDependencies(context.Background(), layer3)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("multiple dependencies", func(t *testing.T) {
		layer4 := &model.Layer{Name: "layer4", Dependencies: []string{"layer2", "layer3"}}
		dependencies, err := stateBackend.ResolveDependencies(context.Background(), layer4)
		require.NoError(t, err)
		assert.Len(t, dependencies, 2)
		assert.Equal(t, layer2, dependencies[0])
		assert.Equal(t, layer3, dependencies[1])
	})
}

func TestFileBackend_ListLayers(t *testing.T) {
	layer1 := &model.Layer{Name: "layer1"}
	layer2 := &model.Layer{Name: "layer2"}
	layer3 := &model.Layer{Name: "layer3"}

	stateBackend := setup(t, []*model.Layer{layer1, layer2, layer3})

	t.Run("list all layers", func(t *testing.T) {
		list, err := stateBackend.ListLayers(context.Background())
		assert.NoError(t, err)
		assert.Len(t, list, 3)
		assert.Contains(t, list, layer1)
		assert.Contains(t, list, layer2)
		assert.Contains(t, list, layer3)
	})

	t.Run("empty list", func(t *testing.T) {
		emptyBackend := setup(t, []*model.Layer{})
		list, err := emptyBackend.ListLayers(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, list)
	})
}