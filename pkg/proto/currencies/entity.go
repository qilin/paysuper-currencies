package currencies

// GetCorrectionValue returns correction value for passed pair of default one
func (r *CorrectionRule) GetCorrectionValue(pair string) float64 {
	if pair == "" {
		return r.CommonCorrection
	}
	if len(pair) != 6 {
		return 0
	}
	v, ok := r.PairCorrection[pair]
	if ok {
		return v
	}
	return r.CommonCorrection
}
