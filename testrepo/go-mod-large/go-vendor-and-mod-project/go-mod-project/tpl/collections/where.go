// Copyright 2017 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collections

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Where returns a filtered subset of a given data type.
func (ns *Namespace) Where(seq, key interface{}, args ...interface{}) (interface{}, error) {
	seqv, isNil := indirect(reflect.ValueOf(seq))
	if isNil {
		return nil, errors.New("can't iterate over a nil value of type " + reflect.ValueOf(seq).Type().String())
	}

	mv, op, err := parseWhereArgs(args...)
	if err != nil {
		return nil, err
	}

	var path []string
	kv := reflect.ValueOf(key)
	if kv.Kind() == reflect.String {
		path = strings.Split(strings.Trim(kv.String(), "."), ".")
	}

	switch seqv.Kind() {
	case reflect.Array, reflect.Slice:
		return ns.checkWhereArray(seqv, kv, mv, path, op)
	case reflect.Map:
		return ns.checkWhereMap(seqv, kv, mv, path, op)
	default:
		return nil, fmt.Errorf("can't iterate over %v", seq)
	}
}

func (ns *Namespace) checkCondition(v, mv reflect.Value, op string) (bool, error) {
	v, vIsNil := indirect(v)
	if !v.IsValid() {
		vIsNil = true
	}

	mv, mvIsNil := indirect(mv)
	if !mv.IsValid() {
		mvIsNil = true
	}
	if vIsNil || mvIsNil {
		switch op {
		case "", "=", "==", "eq":
			return vIsNil == mvIsNil, nil
		case "!=", "<>", "ne":
			return vIsNil != mvIsNil, nil
		}
		return false, nil
	}

	if v.Kind() == reflect.Bool && mv.Kind() == reflect.Bool {
		switch op {
		case "", "=", "==", "eq":
			return v.Bool() == mv.Bool(), nil
		case "!=", "<>", "ne":
			return v.Bool() != mv.Bool(), nil
		}
		return false, nil
	}

	var ivp, imvp *int64
	var fvp, fmvp *float64
	var svp, smvp *string
	var slv, slmv interface{}
	var ima []int64
	var fma []float64
	var sma []string
	if mv.Type() == v.Type() {
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			iv := v.Int()
			ivp = &iv
			imv := mv.Int()
			imvp = &imv
		case reflect.String:
			sv := v.String()
			svp = &sv
			smv := mv.String()
			smvp = &smv
		case reflect.Float64:
			fv := v.Float()
			fvp = &fv
			fmv := mv.Float()
			fmvp = &fmv
		case reflect.Struct:
			switch v.Type() {
			case timeType:
				iv := toTimeUnix(v)
				ivp = &iv
				imv := toTimeUnix(mv)
				imvp = &imv
			}
		case reflect.Array, reflect.Slice:
			slv = v.Interface()
			slmv = mv.Interface()
		}
	} else {
		if mv.Kind() != reflect.Array && mv.Kind() != reflect.Slice {
			return false, nil
		}

		if mv.Len() == 0 {
			return false, nil
		}

		if v.Kind() != reflect.Interface && mv.Type().Elem().Kind() != reflect.Interface && mv.Type().Elem() != v.Type() && v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
			return false, nil
		}
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			iv := v.Int()
			ivp = &iv
			for i := 0; i < mv.Len(); i++ {
				if anInt, err := toInt(mv.Index(i)); err == nil {
					ima = append(ima, anInt)
				}
			}
		case reflect.String:
			sv := v.String()
			svp = &sv
			for i := 0; i < mv.Len(); i++ {
				if aString, err := toString(mv.Index(i)); err == nil {
					sma = append(sma, aString)
				}
			}
		case reflect.Float64:
			fv := v.Float()
			fvp = &fv
			for i := 0; i < mv.Len(); i++ {
				if aFloat, err := toFloat(mv.Index(i)); err == nil {
					fma = append(fma, aFloat)
				}
			}
		case reflect.Struct:
			switch v.Type() {
			case timeType:
				iv := toTimeUnix(v)
				ivp = &iv
				for i := 0; i < mv.Len(); i++ {
					ima = append(ima, toTimeUnix(mv.Index(i)))
				}
			}
		case reflect.Array, reflect.Slice:
			slv = v.Interface()
			slmv = mv.Interface()
		}
	}

	switch op {
	case "", "=", "==", "eq":
		switch {
		case ivp != nil && imvp != nil:
			return *ivp == *imvp, nil
		case svp != nil && smvp != nil:
			return *svp == *smvp, nil
		case fvp != nil && fmvp != nil:
			return *fvp == *fmvp, nil
		}
	case "!=", "<>", "ne":
		switch {
		case ivp != nil && imvp != nil:
			return *ivp != *imvp, nil
		case svp != nil && smvp != nil:
			return *svp != *smvp, nil
		case fvp != nil && fmvp != nil:
			return *fvp != *fmvp, nil
		}
	case ">=", "ge":
		switch {
		case ivp != nil && imvp != nil:
			return *ivp >= *imvp, nil
		case svp != nil && smvp != nil:
			return *svp >= *smvp, nil
		case fvp != nil && fmvp != nil:
			return *fvp >= *fmvp, nil
		}
	case ">", "gt":
		switch {
		case ivp != nil && imvp != nil:
			return *ivp > *imvp, nil
		case svp != nil && smvp != nil:
			return *svp > *smvp, nil
		case fvp != nil && fmvp != nil:
			return *fvp > *fmvp, nil
		}
	case "<=", "le":
		switch {
		case ivp != nil && imvp != nil:
			return *ivp <= *imvp, nil
		case svp != nil && smvp != nil:
			return *svp <= *smvp, nil
		case fvp != nil && fmvp != nil:
			return *fvp <= *fmvp, nil
		}
	case "<", "lt":
		switch {
		case ivp != nil && imvp != nil:
			return *ivp < *imvp, nil
		case svp != nil && smvp != nil:
			return *svp < *smvp, nil
		case fvp != nil && fmvp != nil:
			return *fvp < *fmvp, nil
		}
	case "in", "not in":
		var r bool
		switch {
		case ivp != nil && len(ima) > 0:
			r = ns.In(ima, *ivp)
		case fvp != nil && len(fma) > 0:
			r = ns.In(fma, *fvp)
		case svp != nil:
			if len(sma) > 0 {
				r = ns.In(sma, *svp)
			} else if smvp != nil {
				r = ns.In(*smvp, *svp)
			}
		default:
			return false, nil
		}
		if op == "not in" {
			return !r, nil
		}
		return r, nil
	case "intersect":
		r, err := ns.Intersect(slv, slmv)
		if err != nil {
			return false, err
		}

		if reflect.TypeOf(r).Kind() == reflect.Slice {
			s := reflect.ValueOf(r)

			if s.Len() > 0 {
				return true, nil
			}
			return false, nil
		}
		return false, errors.New("invalid intersect values")
	default:
		return false, errors.New("no such operator")
	}
	return false, nil
}

func evaluateSubElem(obj reflect.Value, elemName string) (reflect.Value, error) {
	if !obj.IsValid() {
		return zero, errors.New("can't evaluate an invalid value")
	}
	typ := obj.Type()
	obj, isNil := indirect(obj)

	// first, check whether obj has a method. In this case, obj is
	// an interface, a struct or its pointer. If obj is a struct,
	// to check all T and *T method, use obj pointer type Value
	objPtr := obj
	if objPtr.Kind() != reflect.Interface && objPtr.CanAddr() {
		objPtr = objPtr.Addr()
	}
	mt, ok := objPtr.Type().MethodByName(elemName)
	if ok {
		if mt.PkgPath != "" {
			return zero, fmt.Errorf("%s is an unexported method of type %s", elemName, typ)
		}
		// struct pointer has one receiver argument and interface doesn't have an argument
		if mt.Type.NumIn() > 1 || mt.Type.NumOut() == 0 || mt.Type.NumOut() > 2 {
			return zero, fmt.Errorf("%s is a method of type %s but doesn't satisfy requirements", elemName, typ)
		}
		if mt.Type.NumOut() == 1 && mt.Type.Out(0).Implements(errorType) {
			return zero, fmt.Errorf("%s is a method of type %s but doesn't satisfy requirements", elemName, typ)
		}
		if mt.Type.NumOut() == 2 && !mt.Type.Out(1).Implements(errorType) {
			return zero, fmt.Errorf("%s is a method of type %s but doesn't satisfy requirements", elemName, typ)
		}
		res := objPtr.Method(mt.Index).Call([]reflect.Value{})
		if len(res) == 2 && !res[1].IsNil() {
			return zero, fmt.Errorf("error at calling a method %s of type %s: %s", elemName, typ, res[1].Interface().(error))
		}
		return res[0], nil
	}

	// elemName isn't a method so next start to check whether it is
	// a struct field or a map value. In both cases, it mustn't be
	// a nil value
	if isNil {
		return zero, fmt.Errorf("can't evaluate a nil pointer of type %s by a struct field or map key name %s", typ, elemName)
	}
	switch obj.Kind() {
	case reflect.Struct:
		ft, ok := obj.Type().FieldByName(elemName)
		if ok {
			if ft.PkgPath != "" && !ft.Anonymous {
				return zero, fmt.Errorf("%s is an unexported field of struct type %s", elemName, typ)
			}
			return obj.FieldByIndex(ft.Index), nil
		}
		return zero, fmt.Errorf("%s isn't a field of struct type %s", elemName, typ)
	case reflect.Map:
		kv := reflect.ValueOf(elemName)
		if kv.Type().AssignableTo(obj.Type().Key()) {
			return obj.MapIndex(kv), nil
		}
		return zero, fmt.Errorf("%s isn't a key of map type %s", elemName, typ)
	}
	return zero, fmt.Errorf("%s is neither a struct field, a method nor a map element of type %s", elemName, typ)
}

// parseWhereArgs parses the end arguments to the where function.  Return a
// match value and an operator, if one is defined.
func parseWhereArgs(args ...interface{}) (mv reflect.Value, op string, err error) {
	switch len(args) {
	case 1:
		mv = reflect.ValueOf(args[0])
	case 2:
		var ok bool
		if op, ok = args[0].(string); !ok {
			err = errors.New("operator argument must be string type")
			return
		}
		op = strings.TrimSpace(strings.ToLower(op))
		mv = reflect.ValueOf(args[1])
	default:
		err = errors.New("can't evaluate the array by no match argument or more than or equal to two arguments")
	}
	return
}

// checkWhereArray handles the where-matching logic when the seqv value is an
// Array or Slice.
func (ns *Namespace) checkWhereArray(seqv, kv, mv reflect.Value, path []string, op string) (interface{}, error) {
	rv := reflect.MakeSlice(seqv.Type(), 0, 0)
	for i := 0; i < seqv.Len(); i++ {
		var vvv reflect.Value
		rvv := seqv.Index(i)
		if kv.Kind() == reflect.String {
			vvv = rvv
			for _, elemName := range path {
				var err error
				vvv, err = evaluateSubElem(vvv, elemName)
				if err != nil {
					continue
				}
			}
		} else {
			vv, _ := indirect(rvv)
			if vv.Kind() == reflect.Map && kv.Type().AssignableTo(vv.Type().Key()) {
				vvv = vv.MapIndex(kv)
			}
		}

		if ok, err := ns.checkCondition(vvv, mv, op); ok {
			rv = reflect.Append(rv, rvv)
		} else if err != nil {
			return nil, err
		}
	}
	return rv.Interface(), nil
}

// checkWhereMap handles the where-matching logic when the seqv value is a Map.
func (ns *Namespace) checkWhereMap(seqv, kv, mv reflect.Value, path []string, op string) (interface{}, error) {
	rv := reflect.MakeMap(seqv.Type())
	keys := seqv.MapKeys()
	for _, k := range keys {
		elemv := seqv.MapIndex(k)
		switch elemv.Kind() {
		case reflect.Array, reflect.Slice:
			r, err := ns.checkWhereArray(elemv, kv, mv, path, op)
			if err != nil {
				return nil, err
			}

			switch rr := reflect.ValueOf(r); rr.Kind() {
			case reflect.Slice:
				if rr.Len() > 0 {
					rv.SetMapIndex(k, elemv)
				}
			}
		case reflect.Interface:
			elemvv, isNil := indirect(elemv)
			if isNil {
				continue
			}

			switch elemvv.Kind() {
			case reflect.Array, reflect.Slice:
				r, err := ns.checkWhereArray(elemvv, kv, mv, path, op)
				if err != nil {
					return nil, err
				}

				switch rr := reflect.ValueOf(r); rr.Kind() {
				case reflect.Slice:
					if rr.Len() > 0 {
						rv.SetMapIndex(k, elemv)
					}
				}
			}
		}
	}
	return rv.Interface(), nil
}

// toFloat returns the float value if possible.
func toFloat(v reflect.Value) (float64, error) {
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	case reflect.Interface:
		return toFloat(v.Elem())
	}
	return -1, errors.New("unable to convert value to float")
}

// toInt returns the int value if possible, -1 if not.
// TODO(bep) consolidate all these reflect funcs.
func toInt(v reflect.Value) (int64, error) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil
	case reflect.Interface:
		return toInt(v.Elem())
	}
	return -1, errors.New("unable to convert value to int")
}

func toUint(v reflect.Value) (uint64, error) {
	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint(), nil
	case reflect.Interface:
		return toUint(v.Elem())
	}
	return 0, errors.New("unable to convert value to uint")
}

// toString returns the string value if possible, "" if not.
func toString(v reflect.Value) (string, error) {
	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Interface:
		return toString(v.Elem())
	}
	return "", errors.New("unable to convert value to string")
}

func toTimeUnix(v reflect.Value) int64 {
	if v.Kind() == reflect.Interface {
		return toTimeUnix(v.Elem())
	}
	if v.Type() != timeType {
		panic("coding error: argument must be time.Time type reflect Value")
	}
	return v.MethodByName("Unix").Call([]reflect.Value{})[0].Int()
}