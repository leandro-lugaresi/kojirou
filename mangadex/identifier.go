package mangadex

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Identifier struct {
	special  bool
	before   int
	after    int
	fallback string
}

func NewIdentifier(id string) Identifier {
	return NewWithFallback(id, id)
}

func UnknownIdentifier() Identifier {
	return Identifier{
		special:  true,
		fallback: "",
	}
}

func NewWithFallback(id string, fallback string) Identifier {
	before, after, ok := parseTwoPart(id)
	switch {
	case ok:
		return Identifier{
			before: before,
			after:  after,
		}
	case fallback == "Unknown":
		return Identifier{
			special: true,
		}
	default:
		return Identifier{
			special:  true,
			fallback: fallback,
		}
	}
}

func (n Identifier) String() string {
	return n.StringFilled(0, 0, false)
}

func (n Identifier) StringFilled(before, after int, forceAfter bool) string {
	switch {
	case n.IsUnknown():
		return "Unknown"
	case n.IsSpecial():
		return n.fallback
	case n.after == 0 && !forceAfter:
		f := fmt.Sprintf("%%0%dd", before)
		return fmt.Sprintf(f, n.before)
	default:
		f := fmt.Sprintf("%%0%dd.%%0%dd", before, after)
		return fmt.Sprintf(f, n.before, n.after)
	}
}

func (n Identifier) Equal(o Identifier) bool {
	switch {
	case !n.IsSpecial() && !o.IsSpecial():
		return n.before == o.before && n.after == o.after
	case !n.IsUnknown() && !o.IsUnknown():
		return n.fallback == o.fallback
	default:
		return false
	}
}

func (n Identifier) Less(o Identifier) bool {
	switch {
	case n.IsUnknown() && o.IsUnknown():
		return false
	case n.IsSpecial() && !o.IsSpecial():
		return false
	case !n.IsSpecial() && o.IsSpecial():
		return true
	case n.IsSpecial() && o.IsSpecial():
		return n.fallback < o.fallback
	case n.before == o.before:
		return n.after < o.after
	default:
		return n.before < o.before
	}
}

func (n Identifier) LessOrEqual(o Identifier) bool {
	return n.Equal(o) || n.Less(o)
}

func (n Identifier) IsSpecial() bool {
	return n.special
}

func (n Identifier) IsUnknown() bool {
	return n.IsSpecial() && len(n.fallback) == 0
}

func (n Identifier) IsNext(o Identifier) bool {
	switch {
	case n.IsSpecial() || o.IsSpecial():
		return true
	case n.before == o.before && n.after < o.after:
		return true
	case n.before+1 == o.before && o.after == 0:
		return true
	default:
		return false
	}
}

func (n Identifier) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}

func (n *Identifier) UnmarshalText(data []byte) error {
	*n = NewWithFallback(string(data), string(data))
	return nil
}

func (n *Identifier) UnmarshalJSON(data []byte) error {
	if string(data) == "nil" {
		*n = UnknownIdentifier()
	}

	text := string("")
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}

	return n.UnmarshalText([]byte(text))
}

func parseTwoPart(s string) (before, after int, ok bool) {
	split := strings.Split(s, ".")
	if len(split) == 0 || len(split) > 2 {
		return 0, 0, false
	} else if len(split) == 1 {
		split = append(split, "0")
	}

	if parsed, err := strconv.ParseUint(split[0], 10, 0); err != nil {
		return 0, 0, false
	} else {
		before = int(parsed)
	}

	if parsed, err := strconv.ParseUint(split[1], 10, 0); err != nil {
		return 0, 0, false
	} else {
		after = int(parsed)
	}

	return before, after, true
}
