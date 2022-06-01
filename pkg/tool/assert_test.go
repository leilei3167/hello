package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFolderFrom(t *testing.T) {
	s, err := GetFolderFrom("https://cloud.tencent.com/developer/article/1808227")
	assert.NoError(t, err)
	t.Log(s)
}
