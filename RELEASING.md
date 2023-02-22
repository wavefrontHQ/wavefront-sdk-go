## How to Release the SDK

1. Make sure that all changes for the release are merged with the master branch.
2. From the root directory of the project in the master branch, run the following commands.
    1. `go install ./...`
    2. `go test ./...`
    3. `go vet ./...`
3. Run `go mod tidy` to make sure it doesn't introduce new changes. If it brings new changes, check those in first before doing the release.
4. Click **Releases** in the about section to get the details on the version history.
5. Decide what the version of the next release will be. We follow [semantic versioning](https://semver.org/).
6. Execute `git tag -a -m 'v0.10.3' 'v0.10.3'` replacing v0.10.3 with the new version.
7. Push the tag up to Github `git push upstream v0.10.3`.
8. Log into Github, click on **Releases** in the about section on the right, and click on **Draft a new release.**
9. For **version**, choose the version you decided upon in step 5.
10. Provide a short but descriptive title for the release.
11. Fill in the details of the release. Please copy the markdown from the previous release and follow the same format.
12. Click "Publish release."
13. From your home directory run `./scripts/add-release-to-pkg-go-dev.sh` to notify pkg.go.dev of the new version.
