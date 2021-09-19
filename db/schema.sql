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
    `owner` int,
    `path` varchar(255),
    `size` int,

    primary key (`id`),
    foreign key (`owner`) references `users`(`id`)
);

create table if not exists `audit_log` (
    `user` int,
    `document` int,
    `action` varchar(64),

    foreign key (`user`) references `users`(`id`),
    foreign key (`document`) references `documents`(`id`)
);