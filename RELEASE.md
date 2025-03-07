# Release Process

This document describes the release process for the Container Device Interface.

1. Create an issue titled `Release container-device-interface vx.y.z` with the following content:
```
- [ ] Create a PR titled `Bump version to vx.y.z` including the following changes:
    - [ ] Change the following to the target version `vx.y.z`:
        - [ ] the `CurrentVersion` in `specs-go/versions.go`,
        - [ ] the `tags.cncf.io/container-device-interface` version in `schema/go.mod`,
        - [ ] (*for specification changes only*) the `tags.cncf.io/container-device-interface/specs-go` version in go.mod,
        - [ ] (*for specification changes only*) the `tags.cncf.io/container-device-interface/specs-go` version in `schema/go.mod`.
    - [ ] Run `make mod-tidy` to update versions in `cmd/**/go.mod`.
    - [ ] Run `make mod-verify` to ensure modules are up to date.
    - [ ] (*for specification changes only*) Add a description to the specification changes in `SPEC.md`.
    - [ ] (*for specification changes only*) Implement a `requiresV*` function for the target version in `specs-go/versions.go`.
- [ ] Merge the PR on sufficient approval.
- [ ] Create a `vx.y.z` tag.
- [ ] (*for specification changes only*) Create a `specs-go/vx.y.w` tag. (for the first spec version `w` will be the same as `z`)
- [ ] Create a GitHub release form the `vx.y.z` tag.
- [ ] (*for specification changes only*) Create a GitHub release from the `specs-go/vx.y.w` tag.
- [ ] Close the release issue.
```
1. Follow the steps as drescribed.
1. If required, create PRs or issues in clients referencing the release issue to update their dependencies.
