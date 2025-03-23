// Copyright (c) 2025 Joshua Sing <joshua@joshuasing.dev>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package golicenser

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// execCommand is exec.Command. It is a function pointer in order to handle exec
// in a reproducible and reliable way in tests runs.
var execCommand = exec.Command

const gitISOTimeFormat = "2006-01-02 15:04:05 -0700"

// gitModRange returns the creation time and last modification time of a file.
func gitModRange(filename string) (time.Time, time.Time, error) {
	// Retrieve file creation time from Git.
	line, err := execCommand("git", "log", "--follow", "--find-renames=70%",
		"--diff-filter=A", "--pretty=format:%cd", "--date=iso", "--", filename).CombinedOutput()
	if err != nil {
		// git log may not have found the commit where the file was added.
		// Instead, retrieve all commits modifying the file and use the time
		// from the first commit.
		line, err = execCommand("git", "log", "--follow", "--find-renames=70%",
			"--reverse", "--pretty=format:%cd", "--date=iso", "--", filename).CombinedOutput()
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("could not get creation time from git: %w", err)
		}
	}
	creationTime, err := time.Parse(gitISOTimeFormat, string(line))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("could not get creation time from git: %w", err)
	}

	// Get file modification time. If the file has been modified locally, this
	// will use the modification time on disk, otherwise the time of the last
	// git commit that modified the file will be used.
	modTime, err := lastModTime(filename)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("could not get modification time: %w", err)
	}

	return creationTime, modTime, nil
}

// gitModTimes returns the times of all commits that modify a file.
func gitModTimes(filename string) ([]time.Time, error) {
	lines, err := execCommand("git", "log", "--follow", "--find-renames=70%", "--diff-filter=ACMR",
		"--reverse", "--pretty=format:%cd", "--date=iso", "--", filename).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("could not get git history: %w", err)
	}

	var modTimes []time.Time
	for _, line := range strings.Split(string(lines), "\n") {
		t, err := time.Parse(gitISOTimeFormat, line)
		if err != nil {
			return nil, fmt.Errorf("could not parse git time %q: %w", line, err)
		}
		modTimes = append(modTimes, t)
	}

	// Check if file has changed locally.
	diff, err := exec.Command("git", "diff", filename).CombinedOutput()
	if err != nil && len(diff) > 0 {
		// File has changed locally, add local modification time.
		fsTime, err := fsModTime(filename)
		if err != nil {
			return nil, fmt.Errorf("could not get fs modification time: %w", err)
		}
		modTimes = append(modTimes, fsTime)
	}

	return modTimes, nil
}

// lastModTime gets the last modification time for a file. It will run
// 'git diff' to determine whether the file has been modified locally, and if
// so, the local file modification time will be returned. Otherwise, the time
// of the last Git commit that modified the file will be returned.
func lastModTime(filename string) (time.Time, error) {
	diff, err := execCommand("git", "diff", filename).CombinedOutput()
	if err == nil && len(diff) == 0 {
		// File has not changed locally, use git commit time.
		return gitModTime(filename)
	}
	return fsModTime(filename)
}

// gitModTime returns the time of the last commit that modified a file.
func gitModTime(filename string) (time.Time, error) {
	line, err := execCommand("git", "log", "-1", "--pretty=format:%cd",
		"--date=iso", "--", filename).CombinedOutput()
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(gitISOTimeFormat, string(line))
}

// fsModTime returns the file modification time from disk.
func fsModTime(filename string) (time.Time, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}
