CREATE SCHEMA if not exists `dbname` DEFAULT CHARACTER SET utf8mb4;

CREATE TABLE dbname.users
(
    id         INT AUTO_INCREMENT PRIMARY KEY,
    uid        VARCHAR(256) NOT NULL unique,
    name       VARCHAR(128) NOT NULL,
    thumbnail  VARCHAR(256),
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_users ON dbname.users (uid);
