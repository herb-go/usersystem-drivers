CREATE TABLE user(
    uid VARCHAR(255) not null,
    created_time BIGINT not null,
    updated_time BIGINT not null,
    status int not null,
    PRIMARY KEY(uid),
    index (created_time,uid)
) DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci  ENGINE=InnoDB; 