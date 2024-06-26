module github.com/eli-l/telegram-bot-api/v7

go 1.21

require github.com/stretchr/testify v1.8.4

retract (
 	v7.0.0 // Missing proper go.mod file
 )

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/mock v0.4.0 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/tools v0.17.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
