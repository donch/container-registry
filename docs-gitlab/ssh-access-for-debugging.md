## SSH access for debugging

Access to the database on staging and production requires access to Teleport.
Follow [this guide](https://gitlab.com/gitlab-com/gl-infra/gitlab-dedicated/sandbox/runbooks/-/blob/master/docs/Teleport/Connect_to_Rails_Console_via_Teleport.md) to install `tsh` and connect to the database.
You may also need an [access request](https://about.gitlab.com/handbook/business-technology/team-member-enablement/onboarding-access-requests/access-requests/#individual-or-bulk-access-request),
you can use [this issue](https://about.gitlab.com/handbook/business-technology/team-member-enablement/onboarding-access-requests/access-requests/#individual-or-bulk-access-request) as example.

### Connecting to the registry DB

Once you have proper access to the SSH hosts, you can follow these steps to connect to the DB

1. Login to Teleport. **Note**: you change the proxy to staging `--proxy=teleport.gprd.gitlab.net`

    ```shell
    tsh login --proxy=teleport.gprd.gitlab.net --request-roles=registry-database-ro --request-reason="Link to issue here with a description"
    ```

2. The step above creates a login request in [#infrastructure-lounge](https://gitlab.slack.com/archives/CB3LSMEJV), wait for approval before continuing.
You can check the status of the request (note the request ID from the login output)

   ```shell
   tsh login --request-id=<request-id>
   ```

3. Login to the DB host

    ```shell
    tsh db login --db-user=console-ro --db-name=gitlabhq_registry db-secondary-registry
    ```

4. SSH into the DB 

    ```shell
    tsh db connect db-secondary-registry
    ```

### Connecting to the Rails console

The [Teleport guide](https://gitlab.com/gitlab-com/gl-infra/gitlab-dedicated/sandbox/runbooks/-/blob/master/docs/Teleport/Connect_to_Rails_Console_via_Teleport.md)
contains the latest steps to do this. This step is to consolidate in one spot for easy reference.

1. Follow [this guide](https://gitlab.com/gitlab-com/gl-infra/gitlab-dedicated/sandbox/runbooks/-/blob/master/docs/bastions/gstg-bastions.md)
to configure the SSH hostnames locally.

2. Login to Teleport and wait for approval (see step 2 above). **Note**: you change the proxy to staging `--proxy=teleport.gprd.gitlab.net`

    ```shell
    tsh login --proxy=teleport.gprd.gitlab.net --request-roles=rails-ro --request-reason="Link to issue here with a description"
    ```

3. SSH into the Rails console

    ```shell
    tsh ssh rails-ro@console-ro-01-sv-gprd
    ```
