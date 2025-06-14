CREATE TABLE IF NOT EXISTS player (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP
);

CREATE TABLE IF NOT EXISTS statistic (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    matches_count INT DEFAULT 0,
    winning_matches INT DEFAULT 0,
    losing_matches INT DEFAULT 0,
    drawn_matches INT DEFAULT 0,
    rating_points INT DEFAULT 0,
    win_rate FLOAT DEFAULT 0.0
);


CREATE TABLE IF NOT EXISTS player_statistic (
    player_id INT UNIQUE,
    statistic_id INT UNIQUE,

    FOREIGN KEY(player_id) REFERENCES player(id) ON DELETE CASCADE,
    FOREIGN KEY(statistic_id) REFERENCES statistic(id) ON DELETE CASCADE
);
