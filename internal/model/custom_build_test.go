package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/windmilleng/tilt/internal/container"
)

func TestAnyFastBuildInfo(t *testing.T) {
	fb := FastBuild{
		BaseDockerfile: "FROM alpine",
		Entrypoint:     Cmd{[]string{"echo", "hi"}},
	}
	cb := CustomBuild{
		Command: "true",
		Deps:    []string{"foo", "bar"},
		Fast:    fb,
	}
	it := ImageTarget{
		BuildDetails: cb,
	}
	bi := it.AnyFastBuildInfo()
	assert.Equal(t, fb, bi)

	it = ImageTarget{
		BuildDetails: fb,
	}
	bi = it.AnyFastBuildInfo()
	assert.Equal(t, fb, bi)

	it = ImageTarget{
		BuildDetails: DockerBuild{},
	}
	bi = it.AnyFastBuildInfo()
	assert.True(t, bi.Empty())
}

func TestEmptyLiveUpdate(t *testing.T) {
	lu, err := NewLiveUpdate(nil, "/base/dir")
	if err != nil {
		t.Fatal(err)
	}
	cb := CustomBuild{
		Command:    "true",
		Deps:       []string{"foo", "bar"},
		LiveUpdate: lu,
	}
	it := ImageTarget{
		BuildDetails: cb,
	}
	bi := it.AnyLiveUpdateInfo()
	assert.True(t, bi.Empty())
}

func TestValidate(t *testing.T) {
	cb := CustomBuild{
		Command: "true",
		Deps:    []string{"foo", "bar"},
	}
	it := NewImageTarget(container.MustParseSelector("gcr.io/foo/bar")).
		WithBuildDetails(cb)

	assert.Nil(t, it.Validate())
}

func TestDoesNotValidate(t *testing.T) {
	cb := CustomBuild{
		Command: "",
		Deps:    []string{"foo", "bar"},
	}
	it := NewImageTarget(container.MustParseSelector("gcr.io/foo/bar")).
		WithBuildDetails(cb)

	assert.Error(t, it.Validate())
}
