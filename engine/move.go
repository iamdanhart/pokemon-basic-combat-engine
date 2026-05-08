package engine

type Category int

const (
	CategoryPhysical Category = iota
	CategorySpecial
	CategoryStatus
)

func (c Category) String() string {
	switch c {
	case CategoryPhysical:
		return "Physical"
	case CategorySpecial:
		return "Special"
	case CategoryStatus:
		return "Status"
	default:
		return "Unknown"
	}
}

type Move struct {
	Name       string
	Type       Type
	Category   Category
	Power      int
	Accuracy   int
	PP         int
	PPMax      int
	EffectFunc string
}
