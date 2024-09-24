CREATE TABLE `pipeline`
(
    `id`         integer,
    `created_at` datetime,
    `updated_at` datetime,
    PRIMARY KEY (`id`)
);
CREATE TABLE `pipeline_execution`
(
    `id`           integer,
    `pipeline_id`  integer,
    `started_at`   datetime,
    `completed_at` datetime,
    `created_at`   datetime,
    `updated_at`   datetime,
    PRIMARY KEY (`id`)
);
CREATE TABLE `job`
(
    `id`                integer,
    `pipeline_id`       integer,
    `name`              text,
    `runs_on`           blob,
    `working_directory` text,
    `env_var`           blob,
    `depends_on`        blob,
    `timeout`           integer,
    `created_at`        datetime,
    `updated_at`        datetime,
    PRIMARY KEY (`id`)
);
CREATE TABLE `job_execution`
(
    `id`           integer,
    `job_id`       integer,
    `status`       integer,
    `reason`       blob,
    `started_at`   datetime,
    `completed_at` datetime,
    `created_at`   datetime,
    `updated_at`   datetime,
    PRIMARY KEY (`id`)
);
CREATE TABLE `step`
(
    `id`                integer,
    `pipeline_id`       integer,
    `job_id`            integer,
    `name`              text,
    `user`              text,
    `container`         text,
    `working_directory` text,
    `commands`          blob,
    `env_var`           blob,
    `depends_on`        blob,
    `created_at`        datetime,
    `updated_at`        datetime,
    PRIMARY KEY (`id`)
);
CREATE TABLE `step_execution`
(
    `id`               integer,
    `job_execution_id` integer,
    `step_id`          integer,
    `status`           integer,
    `exit_code`        integer,
    `started_at`       datetime,
    `completed_at`     datetime,
    `created_at`       datetime,
    `updated_at`       datetime,
    PRIMARY KEY (`id`)
);
CREATE TABLE `job_queue`
(
    `id`               integer,
    `status`           integer,
    `job_execution_id` integer,
    `label`            text,
    `heartbeat`        datetime,
    `created_at`       datetime,
    `updated_at`       datetime,
    PRIMARY KEY (`id`)
)