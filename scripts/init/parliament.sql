CREATE TABLE IF NOT EXISTS members
(
    id    uuid PRIMARY KEY,
    name  varchar UNIQUE NOT NULL,
    party varchar      NOT NULL
);

CREATE TABLE IF NOT EXISTS assembly_program
(
    id varchar PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS proceedings
(
    id          varchar PRIMARY KEY,
    name        varchar NOT NULL,
    date        timestamp with time zone,
    url         varchar NOT NULL,
    transcript  text         NOT NULL,
    attachments varchar[],
    program_id  varchar,
    FOREIGN KEY (program_id)
        REFERENCES assembly_program (id)

);

CREATE TABLE IF NOT EXISTS assembly_votes
(
    id            varchar PRIMARY KEY,
    proceeding_id varchar NOT NULL,
    -- this is json map of member.id -> vote type string
    vote_data     jsonb,
    FOREIGN KEY (proceeding_id)
        REFERENCES proceedings (id)

);
