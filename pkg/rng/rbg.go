package rng

import (
	"math/rand"
)

type BagFactory[BagElement any] func() []*BagElement

type RPG[ElementType any] struct {
	rand       *rand.Rand
	bagFactory BagFactory[ElementType]
	bag        []*ElementType
}

// NewRBG - new random bag generator
func NewRBG[BagElement any](seed int64, bagFactory BagFactory[BagElement]) *RPG[BagElement] {
	r := RPG[BagElement]{
		rand:       rand.New(rand.NewSource(seed)),
		bagFactory: bagFactory,
	}

	r.NextBag()

	return &r
}

func (r *RPG[BagElementType]) NextBag() {
	b := r.bagFactory()

	r.rand.Shuffle(len(b), func(i, j int) {
		b[i], b[j] = b[j], b[i]
	})

	r.bag = b
}

func (r *RPG[BagElementType]) Next() *BagElementType {
	p, b := r.bag[0], r.bag[1:]

	if len(b) == 0 {
		r.NextBag()
	} else {
		r.bag = b
	}

	return p
}
