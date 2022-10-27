package gitclient

import (
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/xanzy/go-gitlab"
)

// Find the largest version
// param ListVersionTag: List of keyed tags is version
// return version string
func GetLatestVersion(listVersionTag *map[string]*gitlab.Tag) string {
	// Return 1.0.0 if list is null
	if len(*listVersionTag) == 0 {
		return "1.0.0"
	}
	maxVer, _ := semver.NewVersion("1.0.0")
	for ver := range *listVersionTag {
		version, _ := semver.NewVersion(ver)
		if version.GreaterThan(maxVer) {
			maxVer = version
		}
	}
	return maxVer.String()
}

// Bump version +1
// param Current: version current to bump
// return version string
func BumpVersion(current string, changelogType map[string][]*gitlab.Commit) string {
	verObj, _ := semver.NewVersion(current)

	if len(changelogType["Major"]) > 0 {
		new := verObj.IncMajor()
		return new.String()
	}
	if len(changelogType["Minor"]) > 0 || len(changelogType["Missing"]) > 0 {
		new := verObj.IncMinor()
		return new.String()
	}
	if len(changelogType["Patch"]) > 0 {
		new := verObj.IncPatch()
		return new.String()
	}
	return current
}

// Return changelog by type change
func GetChangelogType(listCommits []*gitlab.Commit) map[string][]*gitlab.Commit {
	tagType := map[string]string{
		"#breaking": "Major",
		"breaking:": "Major",
		"#major":    "Major",
		"major:":    "Major",
		"#remove":   "Major",
		"remove:":   "Major",
		"#removed":  "Major",
		"removed:":  "Major",
		"#revert":   "Major",
		"revert:":   "Major",
		"#reverted": "Major",
		"reverted:": "Major",
		"#upgrade":  "Major",
		"upgrade:":  "Major",
		"#upgraded": "Major",
		"upgraded:": "Major",
		"#minor":    "Minor",
		"minor:":    "Minor",
		"#change":   "Minor",
		"change:":   "Minor",
		"#changed":  "Minor",
		"changed:":  "Minor",
		"#add":      "Minor",
		"add:":      "Minor",
		"#added":    "Minor",
		"added:":    "Minor",
		"#update":   "Minor",
		"update:":   "Minor",
		"#updated":  "Minor",
		"updated:":  "Minor",
		"#chore":    "Minor",
		"chore:":    "Minor",
		"#patch":    "Patch",
		"patch:":    "Patch",
		"#patched":  "Patch",
		"patched:":  "Patch",
		"#fix":      "Patch",
		"fix:":      "Patch",
		"#fixed":    "Patch",
		"fixed:":    "Patch",
		"#hotfix":   "Patch",
		"hotfix:":   "Patch",
		"#hotfixed": "Patch",
		"hotfixed:": "Patch",
		"#bugfix":   "Patch",
		"bugfix:":   "Patch",
		"#bugfixed": "Patch",
		"bugfixed:": "Patch",
	}

	result := map[string][]*gitlab.Commit{
		"Major":   {},
		"Minor":   {},
		"Patch":   {},
		"Missing": {},
	}

	listCommits = reverseCommits(listCommits)

	for _, commit := range listCommits {
		isMissing := true
		for key := range tagType {
			// Find type change in commit
			if strings.Contains(strings.ToLower(commit.Title), key) {
				result[tagType[key]] = append(result[tagType[key]], commit)
				isMissing = false
				break
			}
		}
		if isMissing {
			result["Missing"] = append(result["Missing"], commit)
		}
	}
	return result
}

// Create changelog description by diff commit
func DescriptionByDiff(changelogType map[string][]*gitlab.Commit, newVersion string) string {
	var changeLogLines []string
	changeLogLines = append(changeLogLines, fmt.Sprintf("# Release version %s", newVersion))

	if len(changelogType["Major"]) > 0 {
		changeLogLines = append(changeLogLines, "## Major changes")
		for _, commit := range changelogType["Major"] {
			changeLogLines = append(changeLogLines, fmt.Sprintf("- %s %s", commit.ShortID, commit.Title))
		}
	}
	if len(changelogType["Minor"]) > 0 {
		changeLogLines = append(changeLogLines, "## Minor changes")
		for _, commit := range changelogType["Minor"] {
			changeLogLines = append(changeLogLines, fmt.Sprintf("- %s %s", commit.ShortID, commit.Title))
		}
	}
	if len(changelogType["Patch"]) > 0 {
		changeLogLines = append(changeLogLines, "## Patches")
		for _, commit := range changelogType["Patch"] {
			changeLogLines = append(changeLogLines, fmt.Sprintf("- %s %s", commit.ShortID, commit.Title))
		}
	}
	if len(changelogType["Missing"]) > 0 {
		changeLogLines = append(changeLogLines, "## Missing definition")
		for _, commit := range changelogType["Missing"] {
			changeLogLines = append(changeLogLines, fmt.Sprintf("- %s %s", commit.ShortID, commit.Title))
		}
	}
	changeLog := strings.Join(changeLogLines, "\n\n")
	return changeLog
}

// Add changelog description to current changlog by diff hotfix commit
func GetHotfixChangeLog(currentChangeLog string, diff *gitlab.Compare) string {
	if diff != nil {
		var hotfixCommits []string
		listCommits := reverseCommits(diff.Commits)
		for _, commit := range listCommits {
			hotfixCommits = append(hotfixCommits, fmt.Sprintf("- %s %s", commit.ShortID, commit.Title))
		}
		hotfixDescription := strings.Join(hotfixCommits, "\n")
		hotfixDescription = fmt.Sprintf("## Hot update (%s)\n", time.Now().Format(time.RFC3339)) + hotfixDescription

		currentChangeLog = currentChangeLog + "\n\n" + hotfixDescription
	}

	return currentChangeLog
}

// Reverses the order of the commit list
func reverseCommits(listCommits []*gitlab.Commit) []*gitlab.Commit {
	for i, j := 0, len(listCommits)-1; i < j; i, j = i+1, j-1 {
		listCommits[i], listCommits[j] = listCommits[j], listCommits[i]
	}
	return listCommits
}
