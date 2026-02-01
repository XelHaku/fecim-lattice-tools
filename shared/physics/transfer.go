package physics

func PolarizationToConductance(P, Ps, Gmin, Gmax float64) float64 {
	if Ps == 0 {
		return (Gmin + Gmax) / 2
	}

	normalizedP := P / Ps

	if normalizedP < -1 {
		normalizedP = -1
	}
	if normalizedP > 1 {
		normalizedP = 1
	}

	return Gmin + (Gmax-Gmin)*(normalizedP+1)/2
}

func ConductanceToPolarization(G, Gmin, Gmax, Ps float64) float64 {
	if Gmax == Gmin {
		return 0
	}

	normalizedG := (G - Gmin) / (Gmax - Gmin)
	normalizedP := 2*normalizedG - 1

	return normalizedP * Ps
}
