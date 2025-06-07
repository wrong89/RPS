CREATE TABLE IF NOT EXISTS player (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    player_id INT NOT NULL REFERENCES player(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);

CREATE TABLE IF NOT EXISTS statistic (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    matches_count INT,
    winning_matches INT,
    losing_matches INT,
    drawn_matches INT,
    rating_points INT,
    win_rate FLOAT
);


CREATE TABLE IF NOT EXISTS player_statistic (
    player_id INT UNIQUE,
    statistic_id INT UNIQUE,

    FOREIGN KEY(player_id) REFERENCES player(id) ON DELETE CASCADE,
    FOREIGN KEY(statistic_id) REFERENCES statistic(id) ON DELETE CASCADE
);
