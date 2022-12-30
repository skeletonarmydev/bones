module skeletonarmydev/bones

go 1.18

require github.com/bones/server/handlers/github v0.0.0

replace github.com/bones/server/handlers/github v0.0.0 => ./handlers/github

require github.com/bones/server/handlers/circleci v0.0.0

replace github.com/bones/server/handlers/circleci v0.0.0 => ./handlers/circleci

replace github.com/bones/server/handlers/aws v0.0.0 => ./handlers/aws

replace github.com/bones/server/common v0.0.0 => ./common

require (
	github.com/bones/server/handlers/aws v0.0.0
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20221026131551-cf6655e29de4 // indirect
	github.com/acomagu/bufpipe v1.0.3 // indirect
	github.com/bones/server/common v0.0.0 // indirect
	github.com/cloudflare/circl v1.1.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.3.1 // indirect
	github.com/go-git/go-git/v5 v5.5.1 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/terraform-exec v0.17.3 // indirect
	github.com/hashicorp/terraform-json v0.14.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/pjbgf/sha1cd v0.2.3 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/skeema/knownhosts v1.1.0 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/zclconf/go-cty v1.11.0 // indirect
	golang.org/x/crypto v0.3.0 // indirect
	golang.org/x/net v0.2.0 // indirect
	golang.org/x/sys v0.2.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)
