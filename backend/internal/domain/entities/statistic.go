package entities

import "context"

type Statistic struct {
	ID            int
	MatchesCount  int
	WiningMatches int
	LosingMatches int
	DrawnMatches  int
	RatingPoints  int
	WinRate       float32
}

type StatisticRepository interface {
	CreateDefaultStatistic(
		ctx context.Context,
	) (int, error)
	GetStatisticByID(
		ctx context.Context,
		id int,
	) (Statistic, error)
}
