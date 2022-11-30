module bones/server/handlers/github

go 1.18

require github.com/bones/server/common v0.0.0

replace github.com/bones/server/common v0.0.0 => ../../common

require (
	github.com/go-git/go-git/v5 v5.4.2
	github.com/hashicorp/hc-install v0.4.0
	github.com/hashicorp/terraform-exec v0.17.3
)

require (
	github.com/acomagu/bufpipe v1.0.3 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.3.1 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/terraform-json v0.14.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/zclconf/go-cty v1.11.0 // indirect
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e // indirect
	golang.org/x/net v0.0.0-20210326060303-6b1517762897 // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)
