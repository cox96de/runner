CREATE TABLE `pipeline`
(
    `id`             integer,
    `app_install_id` integer,
    `repo_owner`     text,
    `repo_name`      text,
    `head_sha`       text,
    `created_at`     datetime,
    `updated_at`     datetime,
    PRIMARY KEY (`id`)
);
CREATE TABLE `job`
(
    `id`                      integer,
    `uid`                     text,
    `name`                    text,
    `steps`                   blob,
    `pipeline_id`             integer,
    `check_run_id`            integer,
    `runner_job_execution_id` integer,
    `created_at`              datetime,
    `updated_at`              datetime,
    PRIMARY KEY (`id`)
)