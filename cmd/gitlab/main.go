package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ducminhgd/go-gitops/pkg/gitclient"
	"github.com/ducminhgd/go-gitops/tools"

	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

var (
	command, host, token, jobToken, ref, version, projectID, sendTo, sendCC, sendBCC, jobName, mode string
	g                                                                                               gitclient.Gitlab
)

const (
	MODE_COMPACT string = "compact"
	MODE_SIMPLE         = "simple"
)

var rootCmd = &cobra.Command{
	Use: "go run main.go COMMAND PROJECT_ID",
	Short: `List of commands (COMMAND):
- create-branch: create release/* branch
- tag: create a tag
- release: run bot create-branch and tag commands
- send: send an email with change log of release

PROJECT_ID: ID of project on Gitlab`,
	Example: `./gitlab tag ${pid} --ref ${target-branch} --version ${desired-version}`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 2 {
			log.Fatal("[GitOps] - Error: getting more than the desired number of arg.")
		}
		if len(args) < 2 {
			log.Fatal("[GitOps] - Error: receiving too little arg.")
		}
		command = args[0]
		projectID = args[1]
		host, _ = cmd.Flags().GetString("host")
		token, _ = cmd.Flags().GetString("token")
		jobToken, _ = cmd.Flags().GetString("job-token")
		ref, _ = cmd.Flags().GetString("ref")
		version, _ = cmd.Flags().GetString("version")
		sendTo, _ = cmd.Flags().GetString("send-to")
		sendCC, _ = cmd.Flags().GetString("send-cc")
		sendBCC, _ = cmd.Flags().GetString("send-bcc")
		jobName, _ = cmd.Flags().GetString("job-name")
		mode, _ = cmd.Flags().GetString("mode")

		in := false
		commandList := []string{"create-branch", "tag", "release", "send"}
		for _, item := range commandList {
			if item == command {
				in = true
				break
			}
		}

		if !in {
			log.Fatalf("[GitOps] - Error: %s not in {create-branch, tag, release, send}.", command)
		}

		modes := []string{MODE_COMPACT, MODE_SIMPLE}
		for _, item := range modes {
			if item == mode {
				in = true
				break
			}
		}

		if !in {
			mode = MODE_COMPACT
		}
	},
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	log.Println("[GitOps] - Running")

	rootCmd.Flags().StringP("host", "", tools.Getenv("GIT_HOST", "https://gitlab.com"), "Git host, if not provided then get GIT_HOST from environment variables.")
	rootCmd.Flags().StringP("job-token", "", tools.Getenv("CI_JOB_TOKEN", ""), "Job Token for Gitlab authentication, if not provided then get CI_JOB_TOKEN from environment variables.")
	rootCmd.Flags().StringP("token", "", tools.Getenv("GIT_PRIVATE_TOKEN", ""), "Token for Gitlab authentication, if not provided then get GIT_PRIVATE_TOKEN from environment variables.")
	rootCmd.Flags().StringP("ref", "", "master", "Git ref name or commit hash")
	rootCmd.Flags().StringP("version", "", "", "Desired version")
	rootCmd.Flags().StringP("send-to", "", "", "Email address to send email to")
	rootCmd.Flags().StringP("send-cc", "", "", "Email address to send CC email to")
	rootCmd.Flags().StringP("send-bcc", "", "", "Email address to send BCC email to")
	rootCmd.Flags().StringP("job-name", "", tools.Getenv("CI_JOB_NAME", ""), "Job name to send email, if not provided then get CI_JOB_NAME from environment variables.")
	rootCmd.Flags().StringP("mode", "", MODE_COMPACT, "Versioning mode.\n'"+MODE_COMPACT+"': no pump up version if a hotfix is merged into a release.\n'"+MODE_SIMPLE+"': pump up version on every release.\nUnknown value will be replaced as default value.")
	err := rootCmd.Execute()
	if err != nil {
		log.Println("Fail to run custom command")
	}
	var gitClient *gitlab.Client
	if jobToken != "" {
		log.Printf("[GitOps] - Using Job token:  %v", err)
		gitClient, err = gitlab.NewJobClient(jobToken, gitlab.WithBaseURL(host))
	} else {
		log.Printf("[GitOps] - Using Private access token:  %v", err)
		gitClient, err = gitlab.NewClient(token, gitlab.WithBaseURL(host))
	}
	if err != nil {
		log.Printf("[GitOps] - Error:  %v", err)
		os.Exit(1)
	}

	g = gitclient.Gitlab{
		Client:    gitClient,
		ProjectID: projectID,
	}

	if command == "create-branch" || command == "release" {
		newVersion, _ := getVersionAndChangeLog(mode)
		branchName := "release/" + newVersion
		log.Println("[GitOps] - Create branch: ", branchName)
		success := g.CreateNewBranch(&branchName, &ref)
		if mode == MODE_COMPACT && !success {
			os.Exit(1)
		}
	}

	if command == "tag" || command == "release" {
		listTags := g.GetListTags()
		listVersionTag := gitclient.GetListVersionTag(listTags)
		newVersion := version
		changeLog := ""
		// If a version exists and versioning mode is "compact" then existed version will be deleted and re-create with new changelog.
		if mode == MODE_COMPACT && listVersionTag[version] != nil {
			latestCommit := g.GetLatestCommit("release/" + version)
			diff := g.GetDiff(listVersionTag[version].Commit.ID, latestCommit.ID, true)
			if len(diff.Commits) == 0 {
				log.Println("[GitOps] - There is no differences between", version, "and", latestCommit)
				os.Exit(1)
			}
			changeLog = gitclient.GetHotfixChangeLog(listVersionTag[version].Release.Description, diff)
			g.DeleteRelease("v" + version)
			g.DeleteTag("v" + version)
		} else {
			versionGet, changelogType := getVersionAndChangeLog(mode)
			newVersion = versionGet
			changeLog = gitclient.DescriptionByDiff(changelogType, newVersion)
		}
		message := ""
		newVersion = "v" + newVersion
		success := g.CreateNewTag(&newVersion, &ref, &message, &changeLog)
		if !success {
			os.Exit(1)
		}
	}

	if command == "send" {
		if version == "" {
			log.Fatal("[GitOps] - Error: version is required")
		}
		env := tools.Getenv("ENV", "Production")
		project, err := g.GetProject()
		tools.CheckFatal(err)
		tag := g.GetTag("v" + version)
		if project == nil || tag == nil {
			log.Println("[GitOps] - Project or Tag does not exit")
			os.Exit(1)
		}

		subject := fmt.Sprintf("[%s][%s] Release version %s (%s)", env, project.PathWithNamespace, tag.Name, tag.Commit.ShortID)
		if jobName != "" {
			subject += fmt.Sprintf(" - Job: %s", jobName)
		}
		body := tag.Release.Description + fmt.Sprintf("\r\n\r\n- Commit hash: %s", tag.Commit.ID)
		sentMail := tools.SendEmail(sendTo, sendCC, sendBCC, body, subject)
		if !sentMail {
			log.Println("[GitOps] - Send email failed")
			os.Exit(1)
		}
	}

	log.Println("[GitOps] - Success")
	os.Exit(0)
}

func getVersionAndChangeLog(mode string) (string, map[string][]*gitlab.Commit) {
	listTags := g.GetListTags()
	if len(listTags) == 0 {
		return "1.0.0", nil
	}
	listVersionTag := gitclient.GetListVersionTag(listTags)
	lastestVersion := gitclient.GetLatestVersion(&listVersionTag)
	diff := g.GetDiff(listVersionTag[lastestVersion].Commit.ID, ref, true)
	if mode != MODE_COMPACT && len(diff.Diffs) == 0 {
		log.Println("[GitOps] - There is no differences between", lastestVersion, "and", ref)
		os.Exit(1)
	}
	changelogType := gitclient.GetChangelogType(diff.Commits)
	newVersion := gitclient.BumpVersion(lastestVersion, changelogType)

	return newVersion, changelogType
}
