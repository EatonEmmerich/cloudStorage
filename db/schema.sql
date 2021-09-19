create table if not exists `users` (
    `id` int not null auto_increment,
    `username` varchar(64),
    `names` varchar(64) character set utf8mb4 collate utf8mb4_bin,
    `email` varchar(64),
    `cell` varchar(64),
    primary key (`id`)
);

create table if not exists `documents` (
    `id` int not null auto_increment,
    `owner` int not null,
    `path` varchar(255) default '' not null,
    `version` int default 0 not null,
    `size` int default 0 not null,
    `media_type` varchar(255) default '' not null,
    `file_name` varchar(255) default '' not null,

    primary key (`id`),
    foreign key (`owner`) references `users`(`id`)
);

create table if not exists `audit_log` (
    `user` int,
    `document` int,
    `action` varchar(64),

    foreign key (`user`) references `users`(`id`) ON DELETE SET NULL ,
    foreign key (`document`) references `documents`(`id`)  ON DELETE SET NULL
);