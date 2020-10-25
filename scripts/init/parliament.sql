CREATE TABLE IF NOT EXISTS members
(
    id    uuid PRIMARY KEY,
    name  varchar UNIQUE NOT NULL,
    party varchar        NOT NULL
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
    transcript  text    NOT NULL,
    attachments varchar[],
    program_id  varchar REFERENCES assembly_program (id)

);

CREATE TABLE IF NOT EXISTS assembly_votes
(
    id    varchar PRIMARY KEY,
    date  timestamp with time zone,
    title text NOT NULL
);

CREATE TABLE IF NOT EXISTS member_votes (
    vote_id varchar REFERENCES assembly_votes NOT NULL,
    member_id uuid REFERENCES members NOT NULL,
    proceedings_id  varchar REFERENCES proceedings NOT NULL,
    vote_type varchar NOT NULL,
    PRIMARY KEY (vote_id, proceedings_id, member_id)
);
