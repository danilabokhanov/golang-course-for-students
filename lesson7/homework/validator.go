package homework

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
)

var ErrNotStruct = errors.New("wrong argument given, should be a struct")
var ErrInvalidValidatorSyntax = errors.New("invalid validator syntax")
var ErrValidateForUnexportedFields = errors.New("validation for unexported field is not allowed")

type ValidationError struct {
	Err error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var message string
	for _, e := range v {
		message += e.Err.Error() + "\n"
	}
	if len(message) > 0 {
		message = message[:len(message)-1]
	}
	return message
}

func (errs ValidationErrors) Is(target error) bool {
	for _, e := range errs {
		if e.Err == target {
			return true
		}
	}
	return false
}

const (
	intID    uint = 2
	stringID uint = 24
	structID uint = 25
)

const (
	lenPref string = "len:"
	inPref  string = "in:"
	minPref string = "min:"
	maxPref string = "max:"
)

const INF int = 1e9

type ValidateParams struct {
	tp            string
	text          []string
	min, max, len int
}

var errs ValidationErrors

func ParseValidateTags(s string, tp reflect.Type) (ValidateParams, error) {
	var objID uint = uint(tp.Kind())
	if objID != intID && objID != stringID {
		return ValidateParams{}, ErrInvalidValidatorSyntax
	}
	opt := ""
	for _, p := range []string{lenPref, inPref, minPref, maxPref} {
		if strings.HasPrefix(s, p) {
			opt = p
			break
		}
	}
	if len(opt) == 0 || (objID != stringID && opt == lenPref) {
		return ValidateParams{}, ErrInvalidValidatorSyntax
	}
	res := ValidateParams{tp: opt, min: -INF, max: INF}
	values := strings.Split(s[len(opt):], ",")
	if opt == inPref {
		if len(values[0]) == 0 {
			return ValidateParams{}, ErrInvalidValidatorSyntax
		}
		if objID == intID {
			for _, t := range values {
				if _, err := strconv.Atoi(t); err != nil {
					return ValidateParams{}, ErrInvalidValidatorSyntax
				}
			}
		}
		res.text = values
	} else {
		num, err := strconv.Atoi(values[0])
		if err != nil {
			return ValidateParams{}, ErrInvalidValidatorSyntax
		}
		if opt == lenPref {
			res.len = num
		} else if opt == minPref {
			res.min = num
		} else {
			res.max = num
		}
	}
	return res, nil
}

func ParseInt(cnfg *ValidateParams, num int, fieldName string) {
	if cnfg.tp == inPref {
		fnd := false
		for _, t := range cnfg.text {
			if x, _ := strconv.Atoi(t); x == num {
				fnd = true
				break
			}
		}
		if !fnd {
			errs = append(errs, ValidationError{fmt.Errorf("wrong field " +
				fieldName + ": the number is not in the list")})
		}
	}
	if cnfg.tp == minPref && num < cnfg.min {
		errs = append(errs, ValidationError{fmt.Errorf("wrong field " +
			fieldName + ": the number is less than the lower bound")})
	}
	if cnfg.tp == maxPref && num > cnfg.max {
		errs = append(errs, ValidationError{fmt.Errorf("wrong field " +
			fieldName + ": the number is greater than the upper bound")})
	}
}

func ParseString(cnfg *ValidateParams, s string, fieldName string) {
	if cnfg.tp == inPref {
		fnd := false
		for _, t := range cnfg.text {
			if strings.Contains(t, s) {
				fnd = true
				break
			}
		}
		if !fnd {
			errs = append(errs, ValidationError{fmt.Errorf("wrong field " +
				fieldName + ": the string is not in the list")})
		}
	}
	if cnfg.tp == lenPref && len(s) != cnfg.len {
		errs = append(errs, ValidationError{fmt.Errorf("wrong field " +
			fieldName + ": incorrect string len")})
	}
	if cnfg.tp == minPref && len(s) < cnfg.min {
		errs = append(errs, ValidationError{fmt.Errorf("wrong field " +
			fieldName + ": the number is less than the lower bound")})
	}
	if cnfg.tp == maxPref && len(s) > cnfg.max {
		errs = append(errs, ValidationError{fmt.Errorf("wrong field " +
			fieldName + ": the number is greater than the upper bound")})
	}
}

func Validate(v any) error {
	val := reflect.ValueOf(v)
	origin := reflect.TypeOf(v).Kind()

	errs = ValidationErrors{}
	if uint(origin) != structID {
		errs = append(errs, ValidationError{ErrNotStruct})
	} else {
		t := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := t.Field(i)
			s, ok := field.Tag.Lookup("validate")
			if !ok {
				continue
			}
			if !field.IsExported() {
				errs = append(errs, ValidationError{ErrValidateForUnexportedFields})
			} else {
				cnfg, err := ParseValidateTags(s, field.Type)
				if err != nil {
					errs = append(errs, ValidationError{err})
				} else if uint(field.Type.Kind()) == intID {
					ParseInt(&cnfg, int(val.Field(i).Int()), field.Name)
				} else {
					ParseString(&cnfg, val.Field(i).String(), field.Name)
				}
			}
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}
