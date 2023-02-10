# PAVE

## Local development

Use `docker compose` to start dependencies, eg:

    docker compose up -d

In case some changes happen in the docker folder, then rebuilding the local dev images might be needed. Use `--build` option to trigger new build, eg:

    docker compose up -d --build

When dependencies are up and running, then use `encore` to run your application.

    encore run

## Cleanup local data

To remove all data, just use the `--volumes` option of `docker compose down`, ie:

    docker compose down --volumes
