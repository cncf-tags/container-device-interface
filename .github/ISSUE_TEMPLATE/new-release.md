---
name: New Release
about: Propose a New Release
title: Release vX.Y.Z
labels: ''
assignees: ''
---

## Release Process

<!--
If making adjustments to the checklist, please also file a PR against
this issue template (.github/ISSUE_TEMPLATE/new-release.md) to update
it accordingly for future releases.
-->

- [ ] Create a PR titled `Bump version to vX.Y.Z` including the following changes:
    - [ ] Change the following to the target version `vX.Y.Z`:
        - [ ] the `tags.cncf.io/container-device-interface` version in `schema/go.mod`,
        - [ ] (*for specification changes only*) the `CurrentVersion` in `specs-go/versions.go`,
        - [ ] (*for specification changes only*) the `tags.cncf.io/container-device-interface/specs-go` version in go.mod,
        - [ ] (*for specification changes only*) the `tags.cncf.io/container-device-interface/specs-go` version in `schema/go.mod`.
    - [ ] Run `make mod-tidy` to update versions in `cmd/**/go.mod`.
    - [ ] Run `make mod-verify` to ensure modules are up to date.
    - [ ] (*for specification changes only*) Add a description to the specification changes in `SPEC.md`.
    - [ ] (*for specification changes only*) Implement a `requiresV*` function for the target version in `specs-go/versions.go`.
- [ ] Merge the PR on sufficient approval.
- [ ] Create a `vX.Y.Z` tag.
- [ ] (*for specification changes only*) Create a `specs-go/vX.Y.W` tag. (for the first spec version `W` will be the same as `Z`)
- [ ] Create a GitHub release form the `vX.Y.Z` tag.
- [ ] (*for specification changes only*) Create a GitHub release from the `specs-go/vX.Y.Z` tag.
- [ ] Close the release issue.
