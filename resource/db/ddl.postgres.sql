CREATE TABLE "pipeline"
(
    "id"         bigserial,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    PRIMARY KEY ("id")
);
CREATE TABLE "pipeline_execution"
(
    "id"           bigserial,
    "pipeline_id"  bigint,
    "started_at"   timestamptz,
    "completed_at" timestamptz,
    "created_at"   timestamptz,
    "updated_at"   timestamptz,
    PRIMARY KEY ("id")
);
CREATE TABLE "job"
(
    "id"                bigserial,
    "pipeline_id"       bigint,
    "name"              text,
    "runs_on"           bytea,
    "working_directory" text,
    "env_var"           bytea,
    "depends_on"        bytea,
    "timeout"           integer,
    "created_at"        timestamptz,
    "updated_at"        timestamptz,
    PRIMARY KEY ("id")
);
CREATE TABLE "job_execution"
(
    "id"           bigserial,
    "job_id"       bigint,
    "status"       integer,
    "reason"       bytea,
    "started_at"   timestamptz,
    "completed_at" timestamptz,
    "created_at"   timestamptz,
    "updated_at"   timestamptz,
    PRIMARY KEY ("id")
);
CREATE TABLE "step"
(
    "id"                bigserial,
    "pipeline_id"       bigint,
    "job_id"            bigint,
    "name"              text,
    "user"              text,
    "container"         text,
    "working_directory" text,
    "commands"          bytea,
    "env_var"           bytea,
    "depends_on"        bytea,
    "created_at"        timestamptz,
    "updated_at"        timestamptz,
    PRIMARY KEY ("id")
);
CREATE TABLE "step_execution"
(
    "id"               bigserial,
    "job_execution_id" bigint,
    "step_id"          bigint,
    "status"           integer,
    "exit_code"        bigint,
    "started_at"       timestamptz,
    "completed_at"     timestamptz,
    "created_at"       timestamptz,
    "updated_at"       timestamptz,
    PRIMARY KEY ("id")
);
CREATE TABLE "job_queue"
(
    "id"               bigserial,
    "status"           integer,
    "job_execution_id" bigint,
    "label"            text,
    "heartbeat"        timestamptz,
    "created_at"       timestamptz,
    "updated_at"       timestamptz,
    PRIMARY KEY ("id")
)