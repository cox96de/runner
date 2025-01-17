CREATE TABLE `pipeline`
(
    `id`             bigint AUTO_INCREMENT,
    `app_install_id` bigint,
    `repo_owner`     longtext,
    `repo_name`      longtext,
    `head_sha`       longtext,
    `created_at`     datetime(3) NULL,
    `updated_at`     datetime(3) NULL,
    PRIMARY KEY (`id`)
);
CREATE TABLE `job`
(
    `id`                      bigint AUTO_INCREMENT,
    `uid`                     longtext,
    `name`                    longtext,
    `steps`                   longblob,
    `pipeline_id`             bigint,
    `check_run_id`            bigint,
    `runner_job_execution_id` bigint,
    `created_at`              datetime(3) NULL,
    `updated_at`              datetime(3) NULL,
    PRIMARY KEY (`id`)
);