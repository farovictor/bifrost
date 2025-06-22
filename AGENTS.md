# Contributor Guidelines for Agents

Automated agents working in this repository must run the following
checks before committing:

```bash
go fmt ./...
go test ./...
```

Only commit if both commands succeed.

Commit messages should be concise and written in the present tense.
Pull request descriptions must summarize the changes and mention test
results.

For local development you can use the provided Makefile:

```bash
make setup
make run
```
