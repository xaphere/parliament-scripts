CREATE TABLE IF NOT EXISTS members
(
    id    uuid PRIMARY KEY,
    name  varchar(100) UNIQUE NOT NULL,
    party varchar(50)         NOT NULL
);

CREATE TABLE IF NOT EXISTS assembly_program
(
    id varchar(32) PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS proceedings
(
    id          varchar(32) PRIMARY KEY,
    name        varchar(128) NOT NULL,
    date        timestamp with time zone,
    url         varchar(128) NOT NULL,
    transcript  text         NOT NULL,
    attachments varchar(128)[],
    program_id  varchar(32),
    FOREIGN KEY (program_id)
        REFERENCES assembly_program (id)

);

CREATE TABLE IF NOT EXISTS assembly_votes
(
    id            varchar(32) PRIMARY KEY,
    proceeding_id varchar(32) NOT NULL,
    -- this is json map of member.id -> vote type string
    vote_data     jsonb,
    FOREIGN KEY (proceeding_id)
        REFERENCES proceedings (id)

);
