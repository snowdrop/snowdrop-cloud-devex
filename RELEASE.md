# Release the project and generate go release

## Make

Execute this command within the root of the project where you pass as parameters your `GITHUB_API_TOKEN` and `VERSION` which corresponds to the tag to be created

```yaml
make upload GITHUB_API_TOKEN=YOURTOKEN VERSION=0.3.0
```

**Remark** : You can next edit the release to add a `changelog` using this command `git log --oneline --decorate v0.2.0..v0.3.0`

## Using goreleaser

Tag the release and push it to the github repo

```bash
git tag -a v0.2.0 -m "Release fixing access to packages files"
git push origin v0.2.0
```

Next, use the [`goreleaser`](https://github.com/goreleaser/goreleaser) tool to build cross platform the project and publish it on github

Create the following `.goreleaser.yml` file
```yaml
builds:
- binary: sb
  env:
    - CGO_ENABLED=0
archive:
  replacements:
    darwin: Darwin
    linux: Linux
    windows: Windows
    386: i386
    amd64: x86_64
checksum:
  name_template: 'checksums.txt'
snapshot:
  name_template: "{{ .Tag }}-next"
changelog:
  sort: asc
  filters:
    exclude:
    - '^docs:'
    - '^test:'
```

Export your `GITHUB_TOKEN` and then execute this command to release

`goreleaser`