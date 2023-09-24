package todotxt

import (
	"fmt"
	"strings"
)

type Priority byte

// Reverse order so that wen can compare using <=
const (
	PrioNone Priority = iota
	PrioZ
	PrioY
	PrioX
	PrioW
	PrioV
	PrioU
	PrioT
	PrioS
	PrioR
	PrioQ
	PrioP
	PrioO
	PrioN
	PrioM
	PrioL
	PrioK
	PrioJ
	PrioI
	PrioH
	PrioG
	PrioF
	PrioE
	PrioD
	PrioC
	PrioB
	PrioA
)

func (p Priority) String() string {
	if p == PrioNone {
		return ""
	}
	return "(" + string(("Z"[0]+1)-byte(p)) + ")"
}

func PriorityFromString(prio string) (Priority, error) {
	prio = strings.ToUpper(strings.TrimSpace(prio))
	if prio == "NONE" {
		return PrioNone, nil
	}
	prio = strings.TrimLeft(prio, "(")
	prio = strings.TrimRight(prio, ")")
	if len(prio) != 1 || prio[0] < "A"[0] || prio[0] > "Z"[0] {
		return 0, fmt.Errorf("priority value \"%s\" out of range (A-Z or NONE)", prio)
	}
	idx := byte(PrioA) - (byte(prio[0]) - "A"[0])
	return Priority(idx), nil
}
