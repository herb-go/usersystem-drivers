CREATE TABLE token(
    uid VARCHAR(255) not null,
    updated_time BIGINT not null,
    token VARCHAR(255),
    PRIMARY KEY(uid)
) DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci ENGINE=InnoDB; 