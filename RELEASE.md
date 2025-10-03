# Provider release process

Our release process relies on [Goreleaser](https://goreleaser.com) for automatically building provider binaries for all architectures, signing them and generating a Github release with the binaries attached.

## Publishing a new version

Once the maintainers are ready to publish a new version, they can create a new git tag starting with `v*` and following [semver](https://semver.org). Pushing this tag will trigger a Github action that runs goreleaser.

They will find a new release with the appropriate version, changelog and attached artifacts on github, that was automatically marked as latest.
