CREATE TABLE `pipeline`
(
    `id`         bigint AUTO_INCREMENT,
    `created_at` datetime(3) NULL,
    `updated_at` datetime(3) NULL,
    PRIMARY KEY (`id`)
);
CREATE TABLE `pipeline_execution`
(
    `id`           bigint AUTO_INCREMENT,
    `pipeline_id`  bigint,
    `started_at`   datetime(3) NULL,
    `completed_at` datetime(3) NULL,
    `created_at`   datetime(3) NULL,
    `updated_at`   datetime(3) NULL,
    PRIMARY KEY (`id`)
);
CREATE TABLE `job`
(
    `id`                bigint AUTO_INCREMENT,
    `pipeline_id`       bigint,
    `name`              longtext,
    `runs_on`           longblob,
    `working_directory` longtext,
    `env_var`           longblob,
    `depends_on`        longblob,
    `created_at`        datetime(3) NULL,
    `updated_at`        datetime(3) NULL,
    PRIMARY KEY (`id`)
);
CREATE TABLE `job_execution`
(
    `id`           bigint AUTO_INCREMENT,
    `job_id`       bigint,
    `status`       int,
    `started_at`   datetime(3) NULL,
    `completed_at` datetime(3) NULL,
    `created_at`   datetime(3) NULL,
    `updated_at`   datetime(3) NULL,
    PRIMARY KEY (`id`)
);
CREATE TABLE `step`
(
    `id`                bigint AUTO_INCREMENT,
    `pipeline_id`       bigint,
    `job_id`            bigint,
    `name`              longtext,
    `user`              longtext,
    `container`         longtext,
    `working_directory` longtext,
    `commands`          longblob,
    `env_var`           longblob,
    `depends_on`        longblob,
    `created_at`        datetime(3) NULL,
    `updated_at`        datetime(3) NULL,
    PRIMARY KEY (`id`)
);
CREATE TABLE `step_execution`
(
    `id`               bigint AUTO_INCREMENT,
    `job_execution_id` bigint,
    `step_id`          bigint,
    `status`           int,
    `exit_code`        int unsigned,
    `started_at`       datetime(3) NULL,
    `completed_at`     datetime(3) NULL,
    `created_at`       datetime(3) NULL,
    `updated_at`       datetime(3) NULL,
    PRIMARY KEY (`id`)
);
