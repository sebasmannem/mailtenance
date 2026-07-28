package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	assert.NotEmpty(t, GetAppVersion())
	assert.Regexp(t, `v(\d+\.)?(\d+\.)?(\*|\d+)$`, GetAppVersion(), "AppVersion should follow convention")
}
