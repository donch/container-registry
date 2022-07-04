<!--
Please use the following format for the issue title:

Release Version vX.Y.Z-gitlab

Example:

Release Version v2.7.7-gitlab
-->

## What's New in this Version
<!--
* Copy the changelog description from https://gitlab.com/gitlab-org/container-registry/-/blob/master/CHANGELOG.md that corresponds to this release, adjusting the headers to `###` for the version diff and `####` for the change categories.

Example:

### [3.43.0](https://gitlab.com/gitlab-org/container-registry/compare/v3.42.0-gitlab...v3.43.0-gitlab) (2022-05-20)


#### Bug Fixes

* gracefully handle missing manifest revisions during imports ([bc7c43f](https://gitlab.com/gitlab-org/container-registry/commit/bc7c43f30d8aba8f2edf2ca741b366614d9234c3))


#### Features

* add ability to check/log whether FIPS crypto has been enabled ([1ac2454](https://gitlab.com/gitlab-org/container-registry/commit/1ac2454ac9dc7eeca5d9b555e0f1e6830fa66439))
* add support for additional gardener media types ([10153f8](https://gitlab.com/gitlab-org/container-registry/commit/10153f8df9a147806084aaff0f95a9d9536bbbe5))
-->

## Tasks

All tasks must be completed (in order) for the release to be considered ~"workflow::production".

### 1. Prepare

1. [ ] Set the milestone of this issue to the target GitLab release.
1. [ ] Set the due date of this issue to the 12th of the release month.

<details>
<summary><b>Instructions</b></summary>
The due date is set to the 12th of each month to create a buffer of 5 days before the merge deadline on the 17th. See [Product Development Timeline](https://about.gitlab.com/handbook/engineering/workflow/#product-development-timeline) for more information about the GitLab release timings.
</details>

### 2. Release

Generate a new release ([documentation](https://gitlab.com/gitlab-org/container-registry/-/tree/master/docs-gitlab#releases)).

### 3. Update

1. [ ] Version bump in [CNG](https://gitlab.com/gitlab-org/build/CNG) is automatically done using the dependency bot. An MR should be found open on the [CNG MR page](https://gitlab.com/gitlab-org/build/CNG/-/merge_requests).
1. [ ] Version bumps for specific distribution paths:
   - [ ] Version bump in [Omnibus](https://gitlab.com/gitlab-org/omnibus-gitlab) is automatically done using the dependency bot. An MR should be found open on the [Omnibus MR page](https://gitlab.com/gitlab-org/omnibus-gitlab/-/merge_requests).
   - [ ] Version bump in [Charts](https://gitlab.com/gitlab-org/charts) is automatically done using the dependency bot. An MR should be found open on the [Charts MR page](https://gitlab.com/groups/gitlab-org/charts/-/merge_requests).
   - [ ] Version bump in [K8s Workloads](https://gitlab.com/gitlab-com/gl-infra/k8s-workloads/gitlab-com). This requires two separate MRs, one for pre-production and staging and another for production, which need to be created and merged in this order. Allow enough time between the two to confirm that everything is working as expected in pre-production and staging. For all environments, update `registry_version` under the respective stanza for each environment in [`bases/environments.yaml`](https://gitlab.com/gitlab-com/gl-infra/k8s-workloads/gitlab-com/-/blob/105b865bbd4c4d745452429b0e3d8ff2e4e52080/bases/environments.yaml):
     - [ ] Pre-production and staging
       - [ ] Update `registry_version` under `pre` and `gstg`
       - [ ] Label with: `/label ~"Service::Container Registry" ~"team::delivery" ~"workflow::ready for review"`
       - [ ] Copy the changelog description from https://gitlab.com/gitlab-org/container-registry/-/blob/master/CHANGELOG.md since the last upgrade to the MR description.
       - [ ] Assign to a reviewer
     - [ ] Production
       - [ ] Update `registry_version` under `gprd`
       - [ ] Label with: `/label ~"Service::Container Registry" ~"team::delivery" ~"workflow::ready for review"`
       - [ ] Copy the changelog description from https://gitlab.com/gitlab-org/container-registry/-/blob/master/CHANGELOG.md since the last upgrade to the MR description.
       - [ ] Assign to a reviewer
1. [ ] Version bump for [GDK](https://gitlab.com/gitlab-org/gitlab-development-kit):
   - [ ] Update `"${registry_image:-registry.gitlab.com/gitlab-org/build/cng/gitlab-container-registry:vX.Y.Z-gitlab}"`, which is passed to the `docker run` command in [support/docker-registry](https://gitlab.com/gitlab-org/gitlab-development-kit/-/blob/main/support/docker-registry)
   - [ ] Label with: `/label ~"workflow::ready for review" ~"group::package" ~"devops::package"`
   - [ ] Copy the changelog description from https://gitlab.com/gitlab-org/container-registry/-/blob/master/CHANGELOG.md since the last upgrade to the MR description.
   - [ ] Assign to the reviewer suggested by reviewer roulette

<details>
<summary><b>Instructions</b></summary>

Bump the Container Registry version used in [CNG](https://gitlab.com/gitlab-org/build/CNG), [Omnibus](https://gitlab.com/gitlab-org/omnibus-gitlab), [Charts](https://gitlab.com/gitlab-org/charts) and [K8s Workloads](https://gitlab.com/gitlab-com/gl-infra/k8s-workloads/gitlab-com).

The CNG image is the pre-requisite for the remaining version bumps which may be merged independently from each other. Only CNG and K8s Workloads version bumps are required for a GitLab.com deployment. The deployment is then completed as documented [here](https://gitlab.com/gitlab-com/gl-infra/k8s-workloads/gitlab-com/-/blob/master/DEPLOYMENT.md). Charts and Omnibus version bumps are required for self-managed releases.

Create a merge request for each project which is not being automatically created. Mark parent tasks as completed once the corresponding merge requests are merged.

Version bump merge requests should appear automatically in the `Related merge requests` section of this issue.

Note: According to the [Distribution Team Merge Request Handling](https://about.gitlab.com/handbook/engineering/development/enablement/distribution/merge_requests.html#assigning-merge-requests) documentation, we should not assign merge requests to an individual.

#### Merge Request Template

For consistency, please use the following template for these merge requests:

##### Branch Name

`bump-container-registry-vX-Y-Z-gitlab`

##### Commit Message

```
Bump Container Registry to vX.Y.Z-gitlab

Changelog: changed
```

##### Title

`Bump Container Registry to vX.Y.Z-gitlab`

##### Description

Repeat the version subsection for multiple versions. As an example, to bump to v2.7.7 in a project where the current version is v2.7.5, create an entry for v2.7.6 and v2.7.7.

```md
## vX.Y.Z-gitlab

[Changelog](https://gitlab.com/gitlab-org/container-registry/blob/release/X.Y-gitlab/CHANGELOG.md#vXYZ-gitlab-YYYY-MM-DD)

Related to <!-- link to this release issue -->.
```

</details>

### 4. Complete

1. [ ] Assign label ~"workflow::verification" once all changes have been merged.
1. [ ] Assign label ~"workflow::production" once all changes have been deployed.
1. [ ] Update all related issues, informing that the deploy is complete.
1. [ ] Close this issue.

<details>
<summary><b>Instructions</b></summary>
To see the version deployed in each environment, look at the [Grafana Container Registry dashboard](https://dashboards.gitlab.net/d/registry-pod/registry-pod-info?orgId=1):

![image](/uploads/3fd5b4902472f6cdcc56b9c2d333472f/image.png)

</details>

/label ~"devops::package" ~"group::package" ~"Category:Container Registry" ~golang ~"workflow::scheduling"
