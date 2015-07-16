package gompatible

import (
	"golang.org/x/tools/go/types"
)

// FuncChange represents a change between functions.
type FuncChange struct {
	Before *Func
	After  *Func
}

func (fc FuncChange) ShowBefore() string {
	f := fc.Before
	return f.Package.showASTNode(f.Doc.Decl)
}

func (fc FuncChange) ShowAfter() string {
	f := fc.After
	return f.Package.showASTNode(f.Doc.Decl)
}

func (fc FuncChange) Kind() ChangeKind {
	switch {
	case fc.Before == nil && fc.After == nil:
		// XXX
		return ChangeUnchanged

	case fc.Before == nil:
		return ChangeAdded

	case fc.After == nil:
		return ChangeRemoved

	// We do not use types.Identical as we want to identify functions by their signature; not by the details of
	// parameters or return types, not:
	//   case types.Identical(fc.Before.Types.Type().Underlying(), fc.After.Types.Type().Underlying()):
	// TODO: make structs so
	case types.ObjectString(fc.Before.Types, nil) == types.ObjectString(fc.After.Types, nil):
		return ChangeUnchanged

	case fc.isCompatible():
		return ChangeCompatible

	default:
		return ChangeBreaking
	}
}

// sigParamsCompatible determines if the parameter parts of two signatures of functions are compatible.
// They are compatible if:
// - The number of parameters equal and the types of parameters are compatible for each of them.
// - The latter parameters have exactly one extra parameter which is a variadic parameter.
func sigParamsCompatible(s1, s2 *types.Signature) bool {
	extra := tuplesCompatibleExtra(s1.Params(), s2.Params())

	switch {
	case extra == nil:
		// s2 params is incompatible with s1 params
		return false

	case len(extra) == 0:
		// s2 params is compatible with s1 params
		return true

	case len(extra) == 1:
		// s2 params is compatible with s1 params with an extra variadic arg
		if s1.Variadic() == false && s2.Variadic() == true {
			return true
		}
	}

	return false
}

func sigResultsCompatible(s1, s2 *types.Signature) bool {
	if s1.Results().Len() == 0 {
		return true
	}

	extra := tuplesCompatibleExtra(s1.Results(), s2.Results())

	switch {
	case extra == nil:
		return false
	case len(extra) == 0:
		return true
	}

	return false
}

func tuplesCompatibleExtra(p1, p2 *types.Tuple) []*types.Var {
	len1 := p1.Len()
	len2 := p2.Len()

	if len1 > len2 {
		return nil
	}

	vars := make([]*types.Var, len2-len1)

	for i := 0; i < len2; i++ {
		if i < len1 {
			v1 := p1.At(i)
			v2 := p2.At(i)

			if v1.Type().String() != v2.Type().String() { // FIXME
				return nil
			}
		} else {
			v2 := p2.At(i)
			vars[i-len1] = v2
		}
	}

	return vars
}

func (fc FuncChange) isCompatible() bool {
	if fc.Before == nil || fc.After == nil {
		return false
	}

	typeBefore, typeAfter := fc.Before.Types.Type(), fc.After.Types.Type()
	if typeBefore == nil || typeAfter == nil {
		return false
	}

	sigBefore, sigAfter := typeBefore.(*types.Signature), typeAfter.(*types.Signature)

	if sigParamsCompatible(sigBefore, sigAfter) == false {
		return false
	}

	if sigResultsCompatible(sigBefore, sigAfter) == false {
		return false
	}

	return true
}
