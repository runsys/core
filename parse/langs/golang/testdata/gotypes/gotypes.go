package gotypes

import (
	"cogentcore.org/core/tree"
)

type Ityp int

type Ptyp *string

type PDtyp *Ptyp

type PMtyp *map[string]Ityp

type Mtyp map[string]int

type Sltyp []float32

type Artyp [20]float64

type Sttyp struct {
	Anon1
	tree.Node
	*APtrField
	FieldA         int
	FieldB         Sltyp
	FieldC, FieldD string
}

type Sl2typ []Ityp

type Iftyp interface {
	tree.Node
	LocalIface
	MethA(a, b Ityp) Ptyp
	MethB(a, b Ityp) Ptyp
}

type Dtyp Sttyp

type DPtyp tree.Node

type Futyp func(ab, bb string, st Sttyp, in int, kk tree.Node) PMtyp

type Fustyp func()

type Furvtyp func() (Ityp, int, string)

type Funrtyp func() (err error)

var Ivar Ityp

var Bvar int

var Svar Sttyp

var Svr2, Svr3 Sttyp

var (
	Mvar map[string]float32
	Avar []tree.Node
)

func (st *Sttyp) Method1(para int, meter string) {
	st.FieldA = para
	st.FieldC = meter
	nwvar := st.FieldA
	var nmvar Sttyp
}

func (st *Sttyp) Method2(para int, meter, feter string, starg *Sttyp) (bool, int) {
	st.FieldA = para
	st.FieldC = meter
	return false, -1
}

func (st *Sttyp) MethodNP() {
	st.FieldA = 22
	st.FieldC = "be"
}
