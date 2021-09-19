create table if not exists `users` (
    `id` int not null auto_increment,
    `username` varchar(64) unique,
    `names` varchar(64) character set utf8mb4 collate utf8mb4_bin,
    `email` varchar(64),
    `cell` varchar(64),
    `hash` varchar(255) default '' not null,
    `salt` varchar(255) default '' not null,

    primary key (`id`),
    index (`username`)
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

create table if not exists `permissions` (
    `id` int not null auto_increment,
    `document_id` int not null,
    `user_id` int not null,
    `pemissions` int not null,

    primary key (`id`),
    foreign key (`document_id`) references `users`(`id`),
    foreign key (`user_id`) references `users`(`id`)
);