CREATE TABLE IF NOT EXISTS player (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR(255) UNIQUE,
    password VARCHAR(255)
);

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
