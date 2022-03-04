# GitLab Container Registry HTTP API V1

> NOTE: This specification is not yet implemented. This document currently serves as foundation to support development.

This document is the specification for the new GitLab Container Registry API.

This is not intended to replace the [Docker Registry HTTP API V2](https://docs.docker.com/registry/spec/api/), superseded by the [OCI Distribution Spec](https://github.com/opencontainers/distribution-spec/blob/main/spec.md), which clients use to upload, download and delete images. That will continue to be maintained and available at `/v2/` (documented [here](../docs/spec/api.md)).

This new API intends to provide additional functionality not covered by the `/v2/` to support the development of new features tailored explicitly for GitLab the product. This API requires the new metadata database and will not be implemented for filesystem metadata.

Please note that this is not the [Container Registry API](https://docs.gitlab.com/ee/api/container_registry.html) of GitLab Rails. This is the API of the Container Registry application. Therefore, while most of the *functionality* described here is expected to surface in the former, parity between the two is not a requirement. Similarly, and although we adhere to the same design principles, the *form* of this API is dictated by the features and constraints of the Container Registry, not by those of GitLab Rails.

## Contents

[TOC]

## Overview

### Versioning

The current `/v2/` and other `/vN/` prefixes are reserved for implementing the OCI Distribution Spec. Therefore, this API uses an independent `/gitlab/v1/` prefix for isolation and versioning purposes.

### Operations

A list of methods and URIs are covered in the table below:

| Method | Path                          | Description                                                   |
| ------ | ----------------------------- | ------------------------------------------------------------- |
| `GET`  | `/gitlab/v1/`                    | Check that the registry implements this API specification. |
| `GET`  | `/gitlab/v1/repositories/<path>/` | Obtain details about the repository identified by `path`.  |
| `PUT`  | `/gitlab/v1/import/<path>/` | Move the repository identified by `path` from filesystem metadata to the database.  |
| `GET`  | `/gitlab/v1/import/<path>/` | Query import status of the repository identified by `path`.  |
| `DELETE`  | `/gitlab/v1/import/<path>/` | Cancel import of the repository identified by `path`.  |

By design, any feature that incurs additional processing time, such as query parameters that allow obtaining additional data, is opt-*in*.

### Authentication

The same authentication mechanism is shared by this and the `/v2/` API. Therefore, clients must obtain a JWT token from the GitLab API using the `/jwt/auth` endpoint.

Considering the above, and unless stated otherwise, all `HEAD` and `GET` requests require a token with `pull` permissions for the target repository(ies), `POST`, `PUT`, and `PATCH` requests require  `push` permissions, and `DELETE` requests require `delete` permissions.

Please refer to the original [documentation](https://docs.docker.com/registry/spec/auth/) from Docker for more details about authentication.

### Strict Slash

We enforce a strict slash policy for all endpoints on this API. This means that all paths must end with a forward slash `/`. A `301 Moved Permanently` response will be issued to redirect the client if the request is sent without the trailing slash. The need to maintain the strict slash policy wil be reviewed on [gitlab-org/container-registry#562](https://gitlab.com/gitlab-org/container-registry/-/issues/562).

## Compliance check

Check if the registry implements the specification described in this document.

### Request

```shell
GET /gitlab/v1/
```

#### Example

```shell
curl --header "Authorization: Bearer <token>" "https://registry.gitlab.com/gitlab/v1/
```

### Response

#### Header

| Status Code        | Reason                                                                                                           |
| ------------------ | ---------------------------------------------------------------------------------------------------------------- |
| `200 OK`           | The registry implements this API specification.                                                                  |
| `401 Unauthorized` | The client should take action based on the contents of the `WWW-Authenticate` header and try the endpoint again. |
| Others             | The registry does not implement this API specification.                                                          |

#### Example

```shell
HTTP/1.1 200 OK
Content-Length: 0
Date: Thu, 25 Nov 2021 16:08:59 GMT
```

## Get repository details

Obtain details about a repository.

### Request

```shell
GET /gitlab/v1/repositories/<path>/
```

| Attribute     | Type    | Required | Default | Description                                                  |
| ------------- | ------- | -------- | ------- | ------------------------------------------------------------ |
| `path`        | String  | Yes      |         | The full path of the target repository. Equivalent to the `name` parameter in the `/v2/` API, described in the [OCI Distribution Spec](https://github.com/opencontainers/distribution-spec/blob/main/spec.md). The same pattern validation applies. |
| `size`        | String | No       |   | If the deduplicated size of the repository should be calculated and included in the response.<br />May be set to `self` or `self_with_descendants`. If set to `self`, the returned value is the deduplicated size of the `path` repository. If set to `self_with_descendants`, the returned value is the deduplicated size of the target repository and any others within. An auth token with `pull` permissions for name `<path>/*` is required for the latter. |

#### Example

```shell
curl --header "Authorization: Bearer <token>" "https://registry.gitlab.com/gitlab/v1/repositories/gitlab-org/build/cng/gitlab-container-registry?size=self"
```

### Response

#### Header

| Status Code        | Reason                                                       |
| ------------------ | ------------------------------------------------------------ |
| `200 OK`           | The repository was found. The response body includes the requested details. |
| `401 Unauthorized` | The client should take action based on the contents of the `WWW-Authenticate` header and try the endpoint again. |
| `400 Bad Request`    | The value of the `size` query parameter is invalid.                                |
| `404 Not Found`    | The repository was not found.                                |

#### Body

| Key          | Value                                                        | Type   | Format                              | Condition                                                    |
| ------------ | ------------------------------------------------------------ | ------ | ----------------------------------- | ------------------------------------------------------------ |
| `name`       | The repository name. This is the last segment of the repository path. | String |                                     |                                                              |
| `path`       | The repository path.                                         | String |                                     |                                                              |
| `size_bytes` | The deduplicated size of the repository (and its descendants, if requested and applicable). This is the sum of all unique image layers referenced by at least one tagged manifest, either directly or indirectly (through a tagged manifest list/index). | Number | Bytes                               | Only present if the request query parameter `size` was set. |
| `created_at` | The timestamp at which the repository was created.           | String | ISO 8601 with millisecond precision |                                                              |
| `updated_at` | The timestamp at which the repository details were last updated. | String | ISO 8601 with millisecond precision | Only present if updated at least once.                       |

#### Example

```json
{
  "name": "gitlab-container-registry",
  "path": "gitlab-org/build/cng/gitlab-container-registry",
  "size_bytes": 28673112401,
  "created_at": "2017-10-17T23:11:13.000+05:30",
  "updated_at": "2021-11-25T14:37:49.251+00:00"
}
```

## Import Repository

Move a single repository from filesystem metadata to the database.

Imports are processed asynchronously, the registry will send a notification via
an HTTP request once the import has finished.

Incoming writes to this repository during the import process will follow the old
code path, and will cause the import process to be cancelled.

### Request

```shell
PUT /gitlab/v1/import/<path>/
```

| Attribute     | Type    | Required | Default   | Description                                                  |
| ------------- | ------- | -------- | --------- | ------------------------------------------------------------ |
| `path`        | String  | Yes      |           | The full path of the target repository. Equivalent to the `name` parameter in the `/v2/` API, described in the [OCI Distribution Spec](https://github.com/opencontainers/distribution-spec/blob/main/spec.md). The same pattern validation applies. |
| `import_type` | Bool    | No       | `final`   | When `import_type=pre`, only import manifests and their associated blobs, without importing tags. Once the pre import is complete, performing a final import (default) should take far less time, reducing the amount of time required during which writes will be blocked. |

#### Example

```shell
curl -X PUT --header "Authorization: Bearer <token>" "https://registry.gitlab.com/gitlab/v1/import/gitlab-org/build/cng/gitlab-container-registry/?import_type=pre"
```

#### Authentication

This endpoint requires an auth token with the `registry` resource type, name set to `import`, and the `*` action. Sample:

```json
{
  "access": [
    {
      "actions": [
        "*"
      ],
      "name": "import",
      "type": "registry"
    }
  ],
  // ...
}
```

### Response
#### Header

| Status Code               | Reason                                                       |
| ------------------------- | ------------------------------------------------------------ |
| `200 OK`                  | The repository was already present on the database and does not need to be imported. This repository may have been previously migrated or native to the database. |
| `202 Accepted`            | The import or pre import was successfully started. |
| `401 Unauthorized`        | The client should take action based on the contents of the `WWW-Authenticate` header and try the endpoint again. |
| `404 Not Found`           | The repository was not found. |
| `409 Conflict`            | The repository is already being imported. |
| `424 Failed Dependency`   | The repository failed to pre import. This error only affects the import request when `import_type=final`, when `import_type=pre` the pre import will be retried.|
| `425 Too Early`           | The Repository is currently being pre imported. |
| `429 Too Many Requests`   | The registry is already running the maximum configured import jobs. |

### Import Notification

Once an import completes, the registry will send a synchronous notification in the form of an HTTP `PUT` request to the endpoint configured in [`migration`](../docs/configuration.md#migration). This notification includes the status of the import and any relevant details about it (such as the reason for a failure).

#### Request
##### Body

| Key          | Value                                                                  | Type   |
| ------------ | ---------------------------------------------------------------------- | ------ |
| `name`       | The repository name. This is the last segment of the repository path.  | String |
| `path`       | The repository path.                                                   | String |
| `status`     | The status of the completed import.                                    | String |
| `detail`     | A detailed explanation of the status. In case an error occurred, this field will contain the reason. | string |

##### Possible Statuses

| Value | Meaning |
| ----- | ------- |
| `import_complete`  | The import was completed successfully. |
| `pre_import_complete` | The pre-import completed successfully. |
| `import_failed` | The import has failed. |
| `pre_import_failed`    | The pre-import has failed. |
| `import_canceled` | The import was canceled. |
| `pre_import_canceled` | The pre-import was canceled. |

##### Examples

###### Success
```json
{
  "name": "gitlab-container-registry",
  "path": "gitlab-org/build/cng/gitlab-container-registry",
  "status": "import_complete",
  "detail": "final import completed successfully"
}
```

###### Error
```json
{
  "name": "gitlab-container-registry",
  "path": "gitlab-org/build/cng/gitlab-container-registry",
  "status": "import_failed",
  "detail": "importing tags: reading tags: write tcp 172.0.0.1:1234->172.0.0.1:4321: write: broken pipe"
}
```

## Get Repository Import Status

Query import status of a repository.

### Request

```shell
GET /gitlab/v1/import/<path>/
```

| Attribute     | Type    | Required | Default   | Description                                                  |
| ------------- | ------- | -------- | --------- | ------------------------------------------------------------ |
| `path`        | String  | Yes      |           | The full path of the target repository. Equivalent to the `name` parameter in the `/v2/` API, described in the [OCI Distribution Spec](https://github.com/opencontainers/distribution-spec/blob/main/spec.md). The same pattern validation applies. |

#### Authentication

This endpoint requires an auth token with the `registry` resource type, name set to `import`, and the `*` action.

#### Example

```shell
curl --header "Authorization: Bearer <token>" "https://registry.gitlab.com/gitlab/v1/import/gitlab-org/build/cng/gitlab-container-registry/"
```

### Response

#### Header

| Status Code        | Reason                                                       |
| ------------------ | ------------------------------------------------------------ |
| `200 OK`           | The repository was found. The response body includes the requested details. |
| `401 Unauthorized` | The client should take action based on the contents of the `WWW-Authenticate` header and try the endpoint again. |
| `404 Not Found`    | The repository was not found.                                |

#### Body

| Key          | Value                                                        | Type   | 
| ------------ | ------------------------------------------------------------ | ------ | 
| `name`       | The repository name. This is the last segment of the repository path. | String |                                     |                                                              |
| `path`       | The repository path.                                         | String |                                     |                                                              |
| `status`     | The status of the import.                                    | String |
| `detail`     | A detailed explanation of the status. | String |

##### Possible Statuses

| Value | Meaning |
| ----- | ------- |
| `native`  | This repository was originally created on the new platform. No import occurred. |
| `import_in_progress`  | The import is in progress. |
| `import_complete`  | The import was completed successfully. |
| `import_failed` | The import has failed. |
| `pre_import_in_progress`  | The pre-import is in progress. |
| `pre_import_complete` | The pre-import completed successfully. |
| `pre_import_failed`    | The pre-import has failed. |
| `import_canceled` | The import was canceled. |
| `pre_import_canceled` | The pre-import was canceled. |

#### Example

```json
{
  "name": "gitlab-container-registry",
  "path": "gitlab-org/build/cng/gitlab-container-registry",
  "status": "import_in_progress",
  "detail": "import in progress"
}
```

## Cancel Repository Import

Cancel ongoing repository (pre)import.

### Request

```shell
DELETE /gitlab/v1/import/<path>/
```

| Attribute     | Type    | Required | Default   | Description                                                  |
| ------------- | ------- | -------- | --------- | ------------------------------------------------------------ |
| `path`        | String  | Yes      |           | The full path of the target repository. Equivalent to the `name` parameter in the `/v2/` API, described in the [OCI Distribution Spec](https://github.com/opencontainers/distribution-spec/blob/main/spec.md). The same pattern validation applies. |

#### Authentication

This endpoint requires an auth token with the `registry` resource type, name set to `import`, and the `*` action.

#### Example

```shell
curl -X DELETE --header "Authorization: Bearer <token>" "https://registry.gitlab.com/gitlab/v1/import/gitlab-org/build/cng/gitlab-container-registry/"
```

### Response

#### Header

| Status Code        | Reason                                                       |
| ------------------ | ------------------------------------------------------------ |
| `202 Accepted`     | The repository was found and the ongoing (pre)import will be canceled. An async notification will be sent once the cancelation occurs. |
| `400 Bad Request`  | The repository was found, but there is no ongoing (pre)import to cancel.  |
| `401 Unauthorized` | The client should take action based on the contents of the `WWW-Authenticate` header and try the endpoint again. |
| `404 Not Found`    | The repository was not found.                                |

#### Body

The response body is only included in `400 Bad Request` responses, as a way to inform the client about the current import status of the repository.

| Key          | Value                                                        | Type   | 
| ------------ | ------------------------------------------------------------ | ------ | 
| `name`       | The repository name. This is the last segment of the repository path. | String |                                     |                                                              |
| `path`       | The repository path.                                         | String |                                     |                                                              |
| `status`     | The status of the import.                                    | String |
| `detail`     | A detailed explanation of the status. | String |

##### Possible Statuses

| Value | Meaning |
| ----- | ------- |
| `native`  | This repository was originally created on the new platform. No import occurred. |
| `import_complete`  | The import was completed successfully. |
| `import_failed` | The import has failed. |
| `pre_import_complete` | The pre-import completed successfully. |
| `pre_import_failed`    | The pre-import has failed. |
| `import_canceled` | The import was canceled. |
| `pre_import_canceled` | The pre-import was canceled. |

#### Example

```json
{
  "name": "gitlab-container-registry",
  "path": "gitlab-org/build/cng/gitlab-container-registry",
  "status": "import_canceled",
  "detail": "import canceled"
}
```

## Errors

In case of an error, the response body payload (if any) follows the format defined in the
[OCI Distribution Spec](https://github.com/opencontainers/distribution-spec/blob/main/spec.md#error-codes), which is the
same format found on the [V2 API](../docs/spec/api.md#errors):

```json
{
    "errors": [
        {
            "code": "<error identifier, see below>",
            "message": "<message describing condition>",
            "detail": "<unstructured>"
        },
        ...
    ]
}
```

### Codes

The error codes encountered via this API are enumerated in the following table. For consistency, whenever possible,
error codes described in the
[OCI Distribution Spec](https://github.com/opencontainers/distribution-spec/blob/main/spec.md#error-codes) are reused.

|Code|Message|Description|
|----|-------|-----------|
`NAME_INVALID` | `invalid repository name` | Invalid repository name encountered either during manifest validation or any API operation.
`NAME_UNKNOWN` | `repository name not known to registry` | This is returned if the name used during an operation is unknown to the registry.
`UNAUTHORIZED` | `authentication required` | The access controller was unable to authenticate the client. Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate.
`INVALID_QUERY_PARAMETER_VALUE` | `the value of a query parameter is invalid` | The value of a request query parameter is invalid. The error detail identifies the concerning parameter and the list of possible values.

## Changes

### 2022-03-03

- Add cancel repository import operation.

### 2022-01-26

- Add get repository import status endpoint.
- Consolidate statuses across the "get repository import status" endpoint and the sync import notifications. 

### 2022-01-13

- Add errors section.

### 2021-12-17

- Added import repository operation.

### 2021-11-26

- Added compliance check operation.
- Added get repository details operation.
