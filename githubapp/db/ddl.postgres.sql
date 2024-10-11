CREATE TABLE "pipeline"
(
    "id"             bigserial,
    "app_install_id" bigint,
    "repo_owner"     text,
    "repo_name"      text,
    "head_sha"       text,
    "created_at"     timestamptz,
    "updated_at"     timestamptz,
    PRIMARY KEY ("id")
);
CREATE TABLE "job"
(
    "id"                      bigserial,
    "uid"                     text,
    "name"                    text,
    "steps"                   bytea,
    "pipeline_id"             bigint,
    "check_run_id"            bigint,
    "runner_job_execution_id" bigint,
    "created_at"              timestamptz,
    "updated_at"              timestamptz,
    PRIMARY KEY ("id")
)