# Git Operator

## Requirements

Install Golang: > go1.15

# Usage

## `gitlab` command

```
List of commands (COMMAND):
- create-branch: create release/* branch
- tag: create a tag
- release: run bot create-branch and tag commands
- send: send an email with change log of release

PROJECT_ID: ID of project on Gitlab

Usage:
  go run main.go COMMAND PROJECT_ID [flags]

Examples:
./gitlab tag ${pid} --ref ${target-branch} --version ${desired-version}

Flags:
  -h, --help              help for go
      --host string       Git host, if not provided then get GIT_HOST from environment variables. (default "https://gitlab.com")
      --job-name string   Job name to send email, if not provided then get CI_JOB_NAME from environment variables.
      --mode string       Versioning mode.
                          'compact': no pump up version if a hotfix is merged into a release.
                          'simple': pump up version on every release.
                          Unknown value will be replaced as default value. (default "compact")
      --ref string        Git ref name or commit hash (default "master")
      --send-bcc string   Email address to send BCC email to
      --send-cc string    Email address to send CC email to
      --send-to string    Email address to send email to
      --token string      Token for Gitlab authentication, if not provided then get GIT_PRIVATE_TOKEN from environment variables. (default "4y_tVHoT-a44cq_DhD9x")
      --version string    Desired version
```

## Run:

1. Run with Go
```bash
go run main.go tag 278 --host=https://gitlab.com--token=this_is_token --ref=master --version=3.60.15 --send-to=abc@gmail.com --send-cc=def@gmail.com --send-bcc=xyz@gmail.com
```

1. Run with Docker
```bash
docker run --rm image-name gitlab tag 278 --host=https://gitlab.com--token=this_is_token --ref=master --version=3.60.15 --send-to=abc@gmail.com --send-cc=def@gmail.com --send-bcc=xyz@gmail.com
```

## Declare environment variables

```
ENV=
GIT_HOST=
GIT_PRIVATE_TOKEN=

SMTP_SERVER=
SMTP_USERNAME=
SMTP_PASSWORD=
```

## Commit message convention

1. A commit message SHOULD contain a tag:
  - Major tags are `#breaking`, `#major`, `#remove`/`#removed`, `#revert`/`#reverted`, `#upgrade`/`#upgrade`, which changes make current application make it not compatible
  - Minor tags are: `#minor`, `#change`/`#changed`,  `#add`/`#added`, `#update`/`#updated`
  - Patch tags are: `#patch`/`#patched`, `#fix`/`#fixed`, `#hotfix`/`#hotfixed`, `#bugfix`/`#bugfixed`
2. If commit message DOES NOT contain a tag, then consider as `#minor`
3. A tag SHOULD be `#<tagName>` or `<tagName>:`.