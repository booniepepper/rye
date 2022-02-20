package env

import (
	"fmt"
	"sort"
	"strings"
)

/* type Envi interface {
	Get(word int) (Object, bool)
	Set(word int, val Object) Object
} */

// This is experimental env without map for Functions with up to two variables

type EnvR2 struct {
	Var1   Object
	Var2   Object
	parent *RyeCtx
	kind   Word
}

type RyeCtx struct {
	state  map[int]Object
	Parent *RyeCtx
	Kind   Word
	locked bool
}

func NewEnv(par *RyeCtx) *RyeCtx {
	var e RyeCtx
	e.state = make(map[int]Object)
	e.Parent = par
	return &e
}

func (e RyeCtx) Probe(idxs Idxs) string {
	var bu strings.Builder
	bu.WriteString("[Context (" + e.Kind.Probe(idxs) + "): ")
	for k, v := range e.state {
		bu.WriteString(idxs.GetWord(k) + ": " + v.Inspect(idxs) + " ")
	}
	bu.WriteString("]")
	return bu.String()
}

const reset = "\x1b[0m"
const reset2 = "\033[39;49m"

const color_word = "\x1b[38;5;45m"
const color_word2 = "\033[38;5;214m"
const color_num2 = "\033[38;5;202m"
const color_string2 = "\033[38;5;148m"
const color_comment = "\033[38;5;247m"

func (e RyeCtx) Preview(idxs Idxs, filter string) string {
	var bu strings.Builder
	bu.WriteString("Context (" + e.Kind.Probe(idxs) + "):\n")
	arr := make([]string, 0)
	i := 0
	for k, v := range e.state {
		str1 := idxs.GetWord(k)
		if strings.Contains(str1, filter) {
			var color string
			switch idxs.GetWord(v.GetKind()) {
			case "builtin":
				color = color_word2
			case "context":
				color = color_num2
			case "function":
				color = color_word
			default:
				color = color_string2
			}
			arr = append(arr, str1+": "+reset+color_comment+v.Inspect(idxs)+reset+"|||"+color) // idxs.GetWord(v.GetKind()
			// bu.WriteString(" " + idxs.GetWord(k) + ": " + v.Inspect(idxs) + "\n")
			i += 1
		}
	}
	sort.Strings(arr)
	for aa := range arr {
		line := arr[aa]
		pars := strings.Split(line, "|||")
		bu.WriteString(" " + pars[1] + pars[0] + "\n")
	}
	return bu.String()
}

// Type returns the type of the Integer.
func (i RyeCtx) Type() Type {
	return CtxType
}

// Inspect returns a string representation of the Integer.
func (i RyeCtx) Inspect(e Idxs) string {
	return i.Probe(e)
}

func (i RyeCtx) Trace(msg string) {
	fmt.Print(msg + "(env): ")
	//fmt.Println(i.Value)
}

func (i RyeCtx) GetKind() int {
	return i.Kind.Index
}

func (e RyeCtx) GetWords(idxs Idxs) Block {
	objs := make([]Object, len(e.state))
	idx := 0
	for k, _ := range e.state {
		objs[idx] = String{idxs.GetWord(k)}
		idx += 1
	}
	return *NewBlock(*NewTSeries(objs))
}

/*func (e *Env) Get(word int) (*Object, bool) {
	obj, exists := e.state[word]
	// recursively look at outer Environments ...
	// only specific functions should do this and ounly for function values ... but there is only global env maybe
	// this is simple environment setup, but we will for the sake of safety and speed change this probably
	// maybe some caching here ... or we could inject functions directly into locked series like some idea was to avoid variable lookup
	if !exists && e.parent != nil {
		par := *e.parent
		obj1, exists1 := par.Get(word)
		if exists1 {
			obj = *obj1
			exists = exists1
		}
	}
	return &obj, exists
}*/

func (e *RyeCtx) Get(word int) (Object, bool) {
	obj, exists := e.state[word]
	// recursively look at outer Environments ...
	// only specific functions should do this and ounly for function values ... but there is only global env maybe
	// this is simple environment setup, but we will for the sake of safety and speed change this probably
	// maybe some caching here ... or we could inject functions directly into locked series like some idea was to avoid variable lookup
	if !exists && e.Parent != nil {
		par := *e.Parent
		obj1, exists1 := par.Get(word)
		if exists1 {
			obj = obj1
			exists = exists1
		}
	}
	return obj, exists
}

func (e *RyeCtx) Set(word int, val Object) Object {
	e.state[word] = val
	return val
}

type ProgramState struct {
	Ser          TSeries // current block of code
	Res          Object  // result of expression
	Ctx          *RyeCtx // Env object ()
	PCtx         *RyeCtx // Env object () -- pure countext
	Idx          *Idxs   // Idx object (index of words)
	Args         []int   // names of current arguments (indexes of names)
	Gen          *Gen    // map[int]map[int]Object  // list of Generic kinds / code
	Inj          Object  // Injected first value in a block evaluation
	Injnow       bool
	ReturnFlag   bool
	ErrorFlag    bool
	FailureFlag  bool
	ForcedResult Object
	SkipFlag     bool
}

func NewProgramState(ser TSeries, idx *Idxs) *ProgramState {
	ps := ProgramState{
		ser,
		nil,
		NewEnv(nil),
		NewEnv(nil),
		idx,
		make([]int, 6),
		NewGen(), //make(map[int]map[int]Object),
		nil,
		false,
		false,
		false,
		false,
		nil,
		false,
	}
	return &ps
}

func AddToProgramState(ps *ProgramState, ser TSeries, idx *Idxs) *ProgramState {
	ps.Ser = ser
	ps.Res = nil
	ps.Idx = idx
	//ps.Env
	return ps
}

func SetValue(ps *ProgramState, word string, val Object) {
	idx, found := ps.Idx.GetIndex(word)
	if found {
		ps.Ctx.Set(idx, val)
		switch valf := val.(type) {
		case Function:
			if valf.Pure {
				ps.PCtx.Set(idx, val)
			}
		}
	}
}
