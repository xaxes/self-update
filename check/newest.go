package check

import (
	"os"
	"path"
	"sort"

	"github.com/Masterminds/semver"
	"go.uber.org/zap"
)

func isNewer(new, curr *semver.Version) bool {
	if curr.GreaterThan(new) || curr.Equal(new) {
		return false
	}
	return true
}

// newest returns the newest binary.
//
// It does not take into account the commit hash.
func newest(currV string, fs []os.FileInfo) (Candidate, error) {
	curr, err := semver.NewVersion(currV)
	if err != nil {
		return Candidate{}, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return Candidate{}, err
	}

	var newer []Candidate
	for _, f := range fs {
		// FIXME: Potential security vulnerability; research if fpath can be a malicious value.
		fpath := path.Join(cwd, f.Name())

		zap.L().Debug("check version", zap.String("bin", fpath))

		new, err := versionFromBin(fpath)
		if err != nil {
			zap.L().Debug("check version", zap.String("bin", fpath), zap.Error(err))
			continue
		}

		if !isNewer(new, curr) {
			continue
		}

		newer = append(newer, Candidate{fpath, new})
	}
	sort.Sort(byVersion(newer))

	if len(newer) == 0 {
		return Candidate{}, ErrNoCandidate
	}

	return newer[len(newer)-1], nil
}

// NewestCandidate returns the binary with the newest version from `dir`.
//
// 1. Execute `<executable> -version` on every applicable binary in `dir`
// 2. Return path of the binary with the latest version.
func NewestCandidate(dir, currVersion string) (Candidate, error) {
	cs, err := updateCandidatesFromDir(dir)
	if err != nil {
		return Candidate{}, err
	}

	if len(cs) == 0 {
		return Candidate{}, ErrNoCandidate
	}

	new, err := newest(currVersion, cs)
	if err != nil {
		return Candidate{}, err
	}

	return new, nil
}
