package currencyrates

func (r *CorrectionRule) GetCorrectionValue(pair string) float64 {
    v, ok := r.PairCorrection[pair]
    if ok {
        return v
    }
    return r.CommonCorrection
}
