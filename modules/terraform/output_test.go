package terraform

import (
	"testing"

	"github.com/terraform-modules-krish/terratest/modules/files"
	"github.com/stretchr/testify/assert"
)

func TestOutputList(t *testing.T) {
	t.Parallel()

	testFolder, err := files.CopyTerraformFolderToTemp("../../test/fixtures/terraform-output-list", t.Name())
	if err != nil {
		t.Fatal(err)
	}

	options := &Options{
		TerraformDir: testFolder,
	}

	InitAndApply(t, options)
	out := OutputList(t, options, "giant_steps")

	expectedLen := 4
	expectedItem := "John Coltrane"
	expectedArray := []string{"John Coltrane", "Tommy Flanagan", "Paul Chambers", "Art Taylor"}

	assert.Len(t, out, expectedLen, "Output should contain %d items", expectedLen)
	assert.Contains(t, out, expectedItem, "Output should contain %q item", expectedItem)
	assert.Equal(t, out[0], expectedItem, "First item should be %q, got %q", expectedItem, out[0])
	assert.Equal(t, out, expectedArray, "Array %q should match %q", expectedArray, out)
}

func TestOutputNotListError(t *testing.T) {
	t.Parallel()

	testFolder, err := files.CopyTerraformFolderToTemp("../../test/fixtures/terraform-output-list", t.Name())
	if err != nil {
		t.Fatal(err)
	}

	options := &Options{
		TerraformDir: testFolder,
	}

	InitAndApply(t, options)
	_, err = OutputListE(t, options, "not_a_list")

	assert.Error(t, err)
}

func TestOutputMap(t *testing.T) {
	t.Parallel()

	testFolder, err := files.CopyTerraformFolderToTemp("../../test/fixtures/terraform-output-map", t.Name())
	if err != nil {
		t.Fatal(err)
	}

	options := &Options{
		TerraformDir: testFolder,
	}

	InitAndApply(t, options)
	out := OutputMap(t, options, "mogwai")

	t.Log(out)

	expectedLen := 4
	expectedMap := map[string]string{
		"guitar_1": "Stuart Braithwaite",
		"guitar_2": "Barry Burns",
		"bass":     "Dominic Aitchison",
		"drums":    "Martin Bulloch",
	}

	assert.Len(t, out, expectedLen, "Output should contain %d items", expectedLen)
	assert.Equal(t, expectedMap, out, "Map %q should match %q", expectedMap, out)
}

func TestOutputNotMapError(t *testing.T) {
	t.Parallel()

	testFolder, err := files.CopyTerraformFolderToTemp("../../test/fixtures/terraform-output-map", t.Name())
	if err != nil {
		t.Fatal(err)
	}

	options := &Options{
		TerraformDir: testFolder,
	}

	InitAndApply(t, options)
	_, err = OutputMapE(t, options, "not_a_map")

	assert.Error(t, err)
}

func TestOutputsForKeys(t *testing.T) {
	t.Parallel()

	testFolder, err := files.CopyTerraformFolderToTemp("../../test/fixtures/terraform-output-all", t.Name())
	if err != nil {
		t.Fatal(err)
	}

	options := &Options{
		TerraformDir: testFolder,
	}

	keys := []string{"our_star", "stars", "magnitudes"}

	InitAndApply(t, options)
	out := OutputForKeys(t, options, keys)

	expectedLen := 3
	assert.Len(t, out, expectedLen, "Output should contain %d items", expectedLen)

	//String value
	expectedString := "Sun"
	str, ok := out["our_star"].(string)
	assert.True(t, ok, "data type wrong")
	assert.Equal(t, expectedString, str, "String %q should match %q", expectedString, str)

	//List value
	expectedListLen := 3
	outputInterfaceList, ok := out["stars"].([]interface{})
	expectedListItem := "Sirius"
	assert.Len(t, outputInterfaceList, expectedListLen, "Output list should contain %d items", expectedListLen)
	assert.Equal(t, expectedListItem, outputInterfaceList[0].(string), "List Item %q should match %q",
		expectedListItem, outputInterfaceList[0].(string))

	//Map value
	outputInterfaceMap, ok := out["magnitudes"].(map[string]interface{})
	expectedMapLen := 3
	expectedMapItem := -1.46
	assert.Len(t, outputInterfaceMap, expectedMapLen, "Output map should contain %d items", expectedMapLen)
	assert.Equal(t, expectedMapItem, outputInterfaceMap["Sirius"].(float64), "Map Item %q should match %q",
		expectedMapItem, outputInterfaceMap["Sirius"].(float64))

	//Key not in the parameter list
	outputNotPresentMap, ok := out["constellations"].(map[string]interface{})
	assert.False(t, ok)
	assert.Nil(t, outputNotPresentMap)
}

func TestOutputsAll(t *testing.T) {
	t.Parallel()

	testFolder, err := files.CopyTerraformFolderToTemp("../../test/fixtures/terraform-output-all", t.Name())
	if err != nil {
		t.Fatal(err)
	}

	options := &Options{
		TerraformDir: testFolder,
	}

	InitAndApply(t, options)
	out := OutputAll(t, options)

	expectedLen := 4
	assert.Len(t, out, expectedLen, "Output should contain %d items", expectedLen)

	//String Value
	expectedString := "Sun"
	str, ok := out["our_star"].(string)
	assert.True(t, ok, "data type wrong")
	assert.Equal(t, expectedString, str, "String %q should match %q", expectedString, str)

	//List Value
	expectedListLen := 3
	outputInterfaceList, ok := out["stars"].([]interface{})
	expectedListItem := "Betelgeuse"
	assert.Len(t, outputInterfaceList, expectedListLen, "Output list should contain %d items", expectedListLen)
	assert.Equal(t, expectedListItem, outputInterfaceList[2].(string), "List item %q should match %q",
		expectedListItem, outputInterfaceList[0].(string))

	//Map Value
	expectedMapLen := 4
	outputInterfaceMap, ok := out["constellations"].(map[string]interface{})
	expectedMapItem := "Aldebaran"
	assert.Len(t, outputInterfaceMap, expectedMapLen, "Output map should contain 4 items")
	assert.Equal(t, expectedMapItem, outputInterfaceMap["Taurus"].(string), "Map item %q should match %q",
		expectedMapItem, outputInterfaceMap["Taurus"].(string))
}

func TestOutputsForKeysError(t *testing.T) {
	t.Parallel()

	testFolder, err := files.CopyTerraformFolderToTemp("../../test/fixtures/terraform-output-map", t.Name())
	if err != nil {
		t.Fatal(err)
	}

	options := &Options{
		TerraformDir: testFolder,
	}

	InitAndApply(t, options)

	_, err = OutputForKeysE(t, options, []string{"random_key"})

	assert.Error(t, err)
}
