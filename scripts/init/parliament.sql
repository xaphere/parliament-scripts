CREATE TABLE IF NOT EXISTS parliamentary_group
(
    id   integer PRIMARY KEY,
    name varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS constituency
(
    id   integer PRIMARY KEY,
    name varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS members
(
    id           integer PRIMARY KEY,
    name         text                       NOT NULL,
    party        integer
        REFERENCES parliamentary_group (id) NOT NULL,
    constituency integer
        REFERENCES constituency (id)        NOT NULL,
    email        varchar
);

CREATE TABLE IF NOT EXISTS assembly_program
(
    id varchar PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS proceedings
(
    id          integer PRIMARY KEY,
    name        varchar NOT NULL,
    date        timestamp with time zone,
    url         varchar NOT NULL,
    transcript  text    NOT NULL,
    attachments varchar[],
    program_id  varchar REFERENCES assembly_program (id)

);

CREATE TABLE IF NOT EXISTS assembly_votes
(
    id varchar PRIMARY KEY,
    date  timestamp with time zone,
    title text NOT NULL
);

CREATE TABLE IF NOT EXISTS member_votes
(
    vote_id        varchar REFERENCES assembly_votes NOT NULL,
    member_id      integer REFERENCES members(id)           NOT NULL,
    proceedings_id varchar REFERENCES proceedings(id)    NOT NULL,
    vote_type      varchar                           NOT NULL,
    PRIMARY KEY (vote_id, proceedings_id, member_id)
);

