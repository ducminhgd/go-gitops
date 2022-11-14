package gitclient

import (
	"log"
	"strings"

	"github.com/xanzy/go-gitlab"
)

type Gitlab struct {
	Client    *gitlab.Client
	ProjectID string
}

// GetListTags gets list tags of project gitlab
func (g Gitlab) GetListTags() []*gitlab.Tag {
	tags, _, err := g.Client.Tags.ListTags(g.ProjectID, nil)
	if err != nil {
		log.Printf("[Gitops] - Error:  %v", err)
	}
	return tags
}

// CreateNewTag creates tag for project gitlab and also create release for that tag
func (g Gitlab) CreateNewTag(tagName, ref, message, releaseDescription *string) bool {
	createTagOption := gitlab.CreateTagOptions{TagName: tagName, Ref: ref, Message: message, ReleaseDescription: releaseDescription}
	_, _, err := g.Client.Tags.CreateTag(g.ProjectID, &createTagOption)
	if err != nil {
		log.Printf("[Gitops] - Error:  %v", err)
		return false
	}

	createReleaseOption := gitlab.CreateReleaseOptions{
		Name:        tagName,
		TagName:     tagName,
		Ref:         ref,
		Description: releaseDescription,
	}
	_, _, err = g.Client.Releases.CreateRelease(g.ProjectID, &createReleaseOption)
	if err != nil {
		log.Printf("[Gitops] - Error:  %v", err)
		return false
	}

	return true
}

// CreateNewBranch creates branch for project gitlab
// param BranchName: name of branch to create
// Param Ref: branch name to create branch from
func (g Gitlab) CreateNewBranch(branchName, ref *string) bool {
	createBranchOptions := gitlab.CreateBranchOptions{Branch: branchName, Ref: ref}
	_, _, err := g.Client.Branches.CreateBranch(g.ProjectID, &createBranchOptions)
	if err != nil {
		log.Printf("[Gitops] - Error:  %v", err)
		return false
	}
	return true
}

// GetListVersionTag Creates list with key is version and value is tag
func GetListVersionTag(listTag []*gitlab.Tag) map[string]*gitlab.Tag {
	listVersion := make(map[string]*gitlab.Tag)
	for _, tag := range listTag {
		listVersion[strings.ReplaceAll(tag.Name, "v", "")] = tag
	}
	return listVersion
}

// GetLatestCommit Returns latest commit of branch
// param Ref: branch name to get commit from
func (g Gitlab) GetLatestCommit(ref string) *gitlab.Commit {
	b := true
	option := gitlab.ListCommitsOptions{RefName: &ref, All: &b, WithStats: &b, FirstParent: &b}
	commit, _, err := g.Client.Commits.ListCommits(g.ProjectID, &option, nil)
	if err != nil {
		log.Printf("[Gitops] - Error:  %v", err)
	}

	if len(commit) == 0 {
		return nil
	}

	return commit[len(commit)-1]
}

// GetDiff Returns compare of two note
func (g Gitlab) GetDiff(from, to string, straight bool) *gitlab.Compare {
	compareOptions := gitlab.CompareOptions{From: &from, To: &to, Straight: &straight}
	diff, _, err := g.Client.Repositories.Compare(g.ProjectID, &compareOptions)
	if err != nil {
		log.Printf("[Gitops] - Error:  %v", err)
	}
	return diff
}

// GetProject Returns a project gitlab
func (g Gitlab) GetProject() (*gitlab.Project, error) {
	project, _, err := g.Client.Projects.GetProject(g.ProjectID, nil)
	if err != nil {
		log.Printf("[Gitops] - Error:  %v", err)
		return nil, err
	}
	return project, nil
}

// GetTag Returns tag of project gitlab
func (g Gitlab) GetTag(ref string) *gitlab.Tag {
	tag, _, err := g.Client.Tags.GetTag(g.ProjectID, ref)
	if err != nil {
		log.Printf("[Gitops] - Error:  %v", err)
	}
	return tag
}

// DeleteRelease Deletes a release of project
func (g Gitlab) DeleteRelease(tagName string) {
	_, _, err := g.Client.Releases.DeleteRelease(g.ProjectID, tagName)
	if err != nil {
		log.Printf("[Gitops] - Error:  %v", err)
	}
}

// DeleteTag Deletes tag by tag name
func (g Gitlab) DeleteTag(tag string) {
	_, err := g.Client.Tags.DeleteTag(g.ProjectID, tag)
	if err != nil {
		log.Printf("[Gitops] - Error:  %v", err)
	}
}

// GetFile gets raw file from repository
func (g Gitlab) GetFile(filePath, ref string) ([]byte, error) {
	option := gitlab.GetRawFileOptions{Ref: &ref}
	file, _, err := g.Client.RepositoryFiles.GetRawFile(g.ProjectID, filePath, &option)
	if err != nil {
		log.Printf("[Gitops] - Error:  %v", err)
		return nil, err
	}
	return file, nil
}

// CommitFiles commits files to a repository
func (g Gitlab) CommitFiles(files map[string]string, ref string) error {
	commitActions := []*gitlab.CommitActionOptions{}
	for filePath, fileContent := range files {
		action := gitlab.CommitActionOptions{
			Action:   gitlab.FileAction(gitlab.FileUpdate),
			FilePath: &filePath,
			Content:  &fileContent,
		}
		commitActions = append(commitActions, &action)
	}
	commitMessage := "Gitops commit"
	createOption := gitlab.CreateCommitOptions{
		Branch:        &ref,
		CommitMessage: &commitMessage,
		Actions:       commitActions,
	}
	_, _, err := g.Client.Commits.CreateCommit(g.ProjectID, &createOption)
	return err
}
