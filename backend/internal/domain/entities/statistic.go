package entities

type Statistic struct {
	ID            int     `json:"id"`
	MatchesCount  int     `json:"matches_count"`
	WiningMatches int     `json:"winning_matches"`
	LosingMatches int     `json:"losing_matches"`
	DrawnMatches  int     `json:"drawn_matches"`
	RatingPoints  int     `json:"rating_points"`
	WinRate       float32 `json:"win_rate"`
}
