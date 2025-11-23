CREATE TABLE teams
(
    team_name TEXT primary key
);

CREATE TABLE reviewers
(
    user_id   TEXT PRIMARY KEY,
    username  TEXT NOT NULL,
    team_name TEXT references teams (team_name) on delete cascade,
    is_active boolean default true
);

CREATE TABLE pull_requests
(
    pull_request_id   VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id         VARCHAR(255) NOT NULL REFERENCES reviewers (user_id),
    status            VARCHAR(20)  NOT NULL DEFAULT 'OPEN'
);

CREATE TABLE pull_request_reviewers
(
    pull_request_id VARCHAR(255) REFERENCES pull_requests (pull_request_id) ON DELETE CASCADE,
    user_id         VARCHAR(255) REFERENCES reviewers (user_id) ON DELETE CASCADE,
    PRIMARY KEY (pull_request_id, user_id)
);

INSERT INTO teams (team_name)
VALUES ('avito'),
       ('payments');

INSERT INTO reviewers (user_id, username, team_name, is_active)
VALUES ('u1', 'Alice', 'avito', true),
       ('u2', 'Lev', 'avito', true),
       ('u3', 'Bob', 'payments', true),
       ('u4', 'Carol', 'payments', true);

INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status)
VALUES ('pr-1001', 'Add search', 'u1', 'OPEN'),
       ('pr-1002', 'Fix payment bug', 'u3', 'OPEN');

INSERT INTO pull_request_reviewers (pull_request_id, user_id)
VALUES ('pr-1001', 'u2'),
       ('pr-1002', 'u4');
