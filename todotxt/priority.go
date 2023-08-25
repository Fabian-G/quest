package todotxt

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
