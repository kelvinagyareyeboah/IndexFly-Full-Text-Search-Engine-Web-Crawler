ngs))
	dl := float64(idx.DocLengths[docID])
	idf := math.Log((N-df+0.5)/(df+0.5) + 1)
	tfNorm := (tf * (k1 + 1)) / (tf + k1*(1-b+b*dl/idx.AvgDL))
	return idf * tfNorm
}
