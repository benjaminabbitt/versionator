package vcs

import (
	"testing"

	"github.com/benjaminabbitt/versionator/internal/vcs/mock"
	"github.com/golang/mock/gomock"
)

// filteringVCS is a VCS that can leave untracked files out of its dirty count.
type filteringVCS struct {
	*mock.MockVersionControlSystem
	ignore []bool
}

func (f *filteringVCS) SetIgnoreUntracked(ignore bool) { f.ignore = append(f.ignore, ignore) }

func TestVCSRegistry_SetIgnoreUntracked_ReachesEveryFilteringVCS(t *testing.T) {
	ctrl := gomock.NewController(t)

	a := &filteringVCS{MockVersionControlSystem: mock.NewMockVersionControlSystem(ctrl)}
	a.EXPECT().Name().Return("a").AnyTimes()
	b := &filteringVCS{MockVersionControlSystem: mock.NewMockVersionControlSystem(ctrl)}
	b.EXPECT().Name().Return("b").AnyTimes()
	// A VCS without the capability is passed over, not an error.
	plain := mock.NewMockVersionControlSystem(ctrl)
	plain.EXPECT().Name().Return("plain").AnyTimes()

	registry := &VCSRegistry{systems: make(map[string]VersionControlSystem)}
	registry.RegisterVCS(a)
	registry.RegisterVCS(b)
	registry.RegisterVCS(plain)

	registry.SetIgnoreUntracked(true)
	registry.SetIgnoreUntracked(false)

	for name, f := range map[string]*filteringVCS{"a": a, "b": b} {
		if len(f.ignore) != 2 || !f.ignore[0] || f.ignore[1] {
			t.Errorf("%s: expected [true false], got %v", name, f.ignore)
		}
	}
}

func TestSetIgnoreUntracked_GlobalRegistry_ReachesRegisteredVCS(t *testing.T) {
	ctrl := gomock.NewController(t)
	f := &filteringVCS{MockVersionControlSystem: mock.NewMockVersionControlSystem(ctrl)}
	f.EXPECT().Name().Return("filtering-test").AnyTimes()

	RegisterVCS(f)
	t.Cleanup(func() { UnregisterVCS("filtering-test") })

	SetIgnoreUntracked(true)
	t.Cleanup(func() { SetIgnoreUntracked(false) })

	if len(f.ignore) != 1 || !f.ignore[0] {
		t.Errorf("expected [true], got %v", f.ignore)
	}
}
