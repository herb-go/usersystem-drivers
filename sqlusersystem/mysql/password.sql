CREATE TABLE password(
    uid VARCHAR(255) not null,
    hash_method VARCHAR(255),
    salt VARCHAR(255) not null,
    password VARCHAR(255)
    CHARACTER SET utf8 
    COLLATE utf8_bin    
    not null,
    updated_time BIGINT not null,
    PRIMARY KEY(uid)    
) DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci ENGINE=InnoDB; 
