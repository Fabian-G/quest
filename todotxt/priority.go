package todotxt

import (
	"errors"
	"strings"
)

type Priority byte

const (
	PrioNone Priority = iota
	PrioA
	PrioB
	PrioC
	PrioD
	PrioE
	PrioF
	PrioG
	PrioH
	PrioI
	PrioJ
	PrioK
	PrioL
	PrioM
	PrioN
	PrioO
	PrioP
	PrioQ
	PrioR
	PrioS
	PrioT
	PrioU
	PrioV
	PrioW
	PrioX
	PrioY
	PrioZ
)

func (p Priority) String() string {
	if p == PrioNone {
		return ""
	}
	return "(" + string(("A"[0]-1)+byte(p)) + ")"
}

func PriorityFromString(prio string) (Priority, error) {
	prio = strings.TrimLeft(prio, "(")
	prio = strings.TrimRight(prio, ")")
	if len(prio) != 1 || prio[0] < "A"[0] || prio[0] > "Z"[0] {
		return 0, errors.New("expected a priority string in format (A)")
	}
	idx := (byte(prio[0]) - "A"[0]) + 1
	return Priority(idx), nil
}
