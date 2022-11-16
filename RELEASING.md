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
13. From your home directory run `mkdir -p go/src/cmd/scratch`.
14. Run `cd go/src/cmd/scratch`.
15. Run `go mod init`.
16. Run `go get github.com/wavefronthq/wavefront-sdk-go@v0.10.3` replacing v0.10.3 with the version you just published.
17. Step 16 lets `pkg.go.dev` know that there is a new version of the SDK to cache.
18. After 15 minutes, go to [pkg.go.dev](https://pkg.go.dev/github.com/wavefronthq/wavefront-sdk-go) and see if the version upgraded. Once the version upgrades on `pkg.go.dev`, all the go users get the new version you just published.
