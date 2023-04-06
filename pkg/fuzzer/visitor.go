package fuzzer

import "toolkit/pkg/feedback"

type Visitor struct {
	V_corpus *Corpus
	V_cov    *feedback.Cov
}
