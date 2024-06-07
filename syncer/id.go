package syncer

import "strconv"

func SeqIDN2S(seqID uint64) string {
	return strconv.FormatUint(seqID, 36)
}

func SeqIDS2N(seqID string) (uint64, error) {
	return strconv.ParseUint(seqID, 36, 64)
}
