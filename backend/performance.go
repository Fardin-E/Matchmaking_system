package backend

func GetRankScore(tier string, division string, lp int) int {
	baseScores := map[string]int{
		"IRON":        0,
		"BRONZE":      200,
		"SILVER":      400,
		"GOLD":        600,
		"PLATINUM":    800,
		"EMERALD":     1000,
		"DIAMOND":     1200,
		"MASTER":      1400,
		"GRANDMASTER": 1600,
		"CHALLENGER":  1800,
	}

	divisionScore := map[string]int{
		"IV":  0,
		"III": 10,
		"II":  20,
		"I":   30,
	}

	return baseScores[tier] + divisionScore[division] + lp
}
