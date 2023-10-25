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
