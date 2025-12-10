package port

type ReferencePoint uint8

const (
	RpObjectCenter ReferencePoint = iota
	RpFacingSide
	RpFacingCorner
	RpFront
	RpBack
	RpExpectedHitPoint
)

func (r ReferencePoint) String() string {
	switch r {
	case RpObjectCenter:
		return "object center"
	case RpFacingSide:
		return "facing side"
	case RpFacingCorner:
		return "facing corner"
	case RpFront:
		return "front"
	case RpBack:
		return "back"
	case RpExpectedHitPoint:
		return "expected hit point"
	default:
		return "unknown"
	}
}
