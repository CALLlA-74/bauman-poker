package pokergame

import "sort"

type SuitType string

const (
	DIAMONDS SuitType = "DIAMONDS"
	HEARTS   SuitType = "HEARTS"
	CLUBS    SuitType = "CLUBS"
	SPADES   SuitType = "SPADES"
)

type IndexType string

const (
	ACE   IndexType = "ACE"
	_2    IndexType = "2"
	_3    IndexType = "3"
	_4    IndexType = "4"
	_5    IndexType = "5"
	_6    IndexType = "6"
	_7    IndexType = "7"
	_8    IndexType = "8"
	_9    IndexType = "9"
	_10   IndexType = "10"
	JACK  IndexType = "JACK"
	QUEEN IndexType = "QUEEN"
	KING  IndexType = "KING"
)

type CombName string

/*const (
	ROYAL_FLUSH    string = "ROYAL-FLUSH"
	STRAIGHT_FLUSH string = "STRAIGHT-FLUSH"
	KARE           string = "KARE"
	FULL_HOUSE     string = "FULL-HOUSE"
	FLUSH          string = "FLUSH"
	STRAIGHT       string = "STRAIGHT"
	TRIPLE         string = "TRIPLE"
	TWO_PAIR       string = "TWO PAIR"
	PAIR           string = "PAIR"
	HIGHEST_CARD   string = "HIGHEST CARD"
)*/

const (
	ROYAL_FLUSH    string = "РОЯЛ-ФЛЕШ"
	STRAIGHT_FLUSH string = "СТРИТ-ФЛЕШ"
	KARE           string = "КАРЕ"
	FULL_HOUSE     string = "ФУЛЛ-ХАУС"
	FLUSH          string = "ФЛЕШ"
	STRAIGHT       string = "СТРИТ"
	TRIPLE         string = "ТРОЙКА"
	TWO_PAIR       string = "ДВЕ ПАРЫ"
	PAIR           string = "ПАРА"
	HIGHEST_CARD   string = "ВЫСШАЯ КАРТА"
)

type PlayingCard struct {
	CardSuit SuitType  `validate:"required"`
	Index    IndexType `validate:"required"`
	weight   int       `json:"-"`
	forCopy  bool      `json:"-"`
}

type BestComb struct {
	name   string
	cards  *[]*PlayingCard
	weight int
	wCard  int // вес старшей
}

var (
	cardSuits   = [...]SuitType{DIAMONDS, HEARTS, CLUBS, SPADES}
	cardIndexes = [...]IndexType{_2, _3, _4, _5, _6, _7, _8, _9, _10, JACK, QUEEN, KING, ACE}
	idxMap      = map[IndexType]int{_2: 2, _3: 3, _4: 4, _5: 5, _6: 6, _7: 7, _8: 8, _9: 9, _10: 10, JACK: 11, QUEEN: 12, KING: 13, ACE: 14}
)

func NewPlayingCard(suit SuitType, index IndexType) *PlayingCard {
	return &PlayingCard{
		CardSuit: suit,
		Index:    index,
		weight:   idxMap[index],
		forCopy:  false,
	}
}

func GetBestComb(inpCards []*PlayingCard) *BestComb {
	switch len(inpCards) {
	case 2:
		if res := isPair(inpCards); res != nil {
			return res
		}
		return isHighestCard(inpCards)
	case 5:
		return check5CardComb(inpCards)
	case 6:
		return check6CardComb(inpCards)
	case 7:
		return check7CardComb(inpCards)
	default:
		return nil
	}
}

func check7CardComb(inpCards []*PlayingCard) *BestComb {
	var maxComb *BestComb
	maxW := 0
	for idx := 0; idx < len(inpCards); idx++ {
		res := make([]*PlayingCard, 6)
		idx2 := 0
		for i := 0; i < len(inpCards); i++ {
			if idx != i {
				res[idx2] = inpCards[i]
				idx2++
			}
		}
		t := check6CardComb(res)
		if t != nil && t.weight > maxW {
			maxComb = t
			maxW = t.weight
		}
	}
	return maxComb
}

func check6CardComb(inpCards []*PlayingCard) *BestComb {
	var maxComb *BestComb
	maxW := 0
	for idx := 0; idx < len(inpCards); idx++ {
		res := make([]*PlayingCard, 5)
		idx2 := 0
		for i := 0; i < len(inpCards); i++ {
			if idx != i {
				res[idx2] = inpCards[i]
				idx2++
			}
		}
		t := check5CardComb(res)
		if t != nil && t.weight > maxW {
			maxComb = t
			maxW = t.weight
		}
	}
	return maxComb
}

func check5CardComb(inpCards []*PlayingCard) *BestComb {
	if len(inpCards) != 5 {
		return nil
	}

	cards := make([]*PlayingCard, len(inpCards))
	copy(cards, inpCards)

	sort.SliceStable(cards, func(a, b int) bool {
		return cards[a].weight < cards[b].weight
	})

	groups := make(map[IndexType][]*PlayingCard)
	for _, c := range cards {
		groups[c.Index] = append(groups[c.Index], c)
	}

	if res := isRoyalFlush(cards); res != nil {
		return res
	}

	switch len(groups) {
	case 2:
		for _, group := range groups {
			if len(group) == 4 {
				return &BestComb{
					name:   KARE,
					cards:  &group,
					weight: 8,
				}
			}
		}
		return &BestComb{
			name:   FULL_HOUSE,
			cards:  &cards,
			weight: 7,
		}
	case 3:
		for _, group := range groups {
			if len(group) == 3 {
				return &BestComb{
					name:   TRIPLE,
					cards:  &group,
					weight: 4,
				}
			}
		}
		return is2Pair(cards)
	case 4:
		return isPair(cards)
	default:
		flush := isFlush(cards)
		straight := isStraight(cards)
		switch {
		case flush && straight:
			return &BestComb{
				name:   STRAIGHT_FLUSH,
				cards:  &cards,
				weight: 9,
			}
		case flush:
			return &BestComb{
				name:   FLUSH,
				cards:  &cards,
				weight: 6,
			}
		case straight:
			return &BestComb{
				name:   STRAIGHT,
				cards:  &cards,
				weight: 5,
			}
		default:
			return isHighestCard(cards)
		}
	}

}

func isRoyalFlush(inpCards []*PlayingCard) *BestComb {
	if len(inpCards) < 5 {
		return nil
	}

	cards := make([]*PlayingCard, len(inpCards))
	copy(cards, inpCards)

	idxTypes := map[IndexType]int{ACE: 0, KING: 1, QUEEN: 2, JACK: 3, _10: 4}
	typesToIdx := map[int]IndexType{0: ACE, 1: KING, 2: QUEEN, 3: JACK, 4: _10}
	suitTypes := map[SuitType]int{DIAMONDS: 0, HEARTS: 1, CLUBS: 2, SPADES: 3}
	flags := [4][5]bool{}

	for idx := range cards {
		flags[suitTypes[cards[idx].CardSuit]][idxTypes[cards[idx].Index]] = true
	}

	for idx := 0; idx < len(flags); idx++ {
		isRes := true
		for j := 0; j < len(flags[idx]); j++ {
			if !flags[idx][j] {
				isRes = false
				break
			}
		}
		if isRes {
			res := make([]*PlayingCard, 5)
			for i := range flags[idx] {
				res[i] = NewPlayingCard(cardSuits[idx], typesToIdx[i])
			}
			return &BestComb{
				name:   ROYAL_FLUSH,
				cards:  &res,
				weight: 10,
				wCard:  14,
			}
		}
	}
	return nil
}

func isFlush(cards []*PlayingCard) bool {
	suit := cards[0].CardSuit
	for i := 1; i < 5; i++ {
		if cards[i].CardSuit != suit {
			return false
		}
	}
	return true
}

func isStraight(cards []*PlayingCard) bool {
	sorted := make([]*PlayingCard, 5)
	copy(sorted, cards)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Index < sorted[j].Index
	})
	if sorted[0].weight+4 == sorted[4].weight {
		return true
	}
	if sorted[4].weight == 14 && sorted[0].weight == 2 && sorted[3].weight == 5 {
		return true
	}
	return false
}

/*func isStraignt(inpCards []*PlayingCard) *BestComb {
	if len(inpCards) < 5 {
		return nil
	}

	cards := make([]*PlayingCard, len(inpCards))
	copy(cards, inpCards)

	numOfMono := 0
	for idx := 0; idx < len(cards) - 1; idx++ {
		if cards[idx].weight + 1 == cards[idx + 1].weight {
			numOfMono++
		} else {
			numOfMono = 0
		}

	}
}*/

func isTriple(inpCards []*PlayingCard) *BestComb {
	if len(inpCards) < 3 {
		return nil
	}

	cards := make([]*PlayingCard, len(inpCards))
	copy(cards, inpCards)

	for idx := len(cards) - 1; idx > 1; {
		if cards[idx].Index == cards[idx-1].Index && cards[idx-1].Index == cards[idx-2].Index {
			cards[idx].forCopy = true
			cards[idx-1].forCopy = true
			cards[idx-2].forCopy = true
			idx -= 3
			break
		} else {
			idx--
		}
	}

	res := make([]*PlayingCard, 3)
	idx2 := 0
	maxW := 0
	for idx := len(cards) - 1; idx >= 0; idx-- {
		if cards[idx].forCopy {
			cards[idx].forCopy = false
			res[idx2] = cards[idx]
			if res[idx2].weight > maxW {
				maxW = res[idx2].weight
			}
			idx2++
		}
	}

	return &BestComb{
		name:   TRIPLE,
		cards:  &res,
		weight: 4,
		wCard:  maxW,
	}
}

func is2Pair(inpCards []*PlayingCard) *BestComb {
	if len(inpCards) < 4 {
		return nil
	}

	cards := make([]*PlayingCard, len(inpCards))
	copy(cards, inpCards)

	for idx := len(cards) - 1; idx > 0; {
		if cards[idx].Index == cards[idx-1].Index {
			cards[idx].forCopy = true
			cards[idx-1].forCopy = true
			idx -= 2
		} else {
			idx--
		}
	}

	res := make([]*PlayingCard, 4)
	idx2 := 0
	maxW := 0
	for idx := len(cards) - 1; idx >= 0; idx-- {
		if cards[idx].forCopy {
			cards[idx].forCopy = false
			res[idx2] = cards[idx]
			if res[idx2].weight > maxW {
				maxW = res[idx2].weight
			}
			idx2++
		}
	}

	return &BestComb{
		name:   TWO_PAIR,
		cards:  &res,
		weight: 3,
		wCard:  maxW,
	}
}

func isPair(inpCards []*PlayingCard) *BestComb {
	if len(inpCards) < 2 {
		return nil
	}

	if inpCards[0].Index == inpCards[1].Index {
		return &BestComb{
			name:   PAIR,
			cards:  &inpCards,
			weight: 2,
			wCard:  inpCards[0].weight,
		}
	}
	return nil
}

func isHighestCard(inpCards []*PlayingCard) *BestComb {
	cards := make([]*PlayingCard, len(inpCards))
	copy(cards, inpCards)

	sort.SliceStable(cards, func(a, b int) bool {
		return cards[a].weight < cards[b].weight
	})

	res := make([]*PlayingCard, 1)
	res[0] = cards[len(cards)-1]
	return &BestComb{
		name:   HIGHEST_CARD,
		cards:  &res,
		weight: 1,
		wCard:  res[0].weight,
	}
}
