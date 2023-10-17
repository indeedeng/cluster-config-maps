# Release Process

TODO: Migrate these to github actions.

### Prerelease

1. Ensure the [CHANGELOG.md](CHANGELOG.md) is up to date.
2. Bump the [Chart.yaml](deploy/charts/cluster-config-maps/Chart.yaml) `version` or `appVersion` as needed.

### Test the build of cluster-config-maps

1. Build the csi driver: `make docker.build`

### Release cluster-config-maps

1. Build the csi driver: `git tag 0.x.x && git push --tags`

### Release Helm Chart

1. Regenerate the helm chart + docs: `helm.generate`