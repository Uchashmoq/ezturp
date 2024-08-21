package tools

import (
	"fmt"
	"strings"
)

type CommandArgs map[string][]string

func isOpt(s string) bool {
	return strings.HasPrefix(s, "-")
}
func getOpt(s string) string {
	return s[1:]
}

func (ca CommandArgs) ContainsValue(value string) bool {
	for _, vs := range ca {
		for _, v := range vs {
			if v == value {
				return true
			}
		}
	}
	return false
}

func (ca CommandArgs) Get(opt string) (vs []string) {
	vs, ok := ca[opt]
	if !ok {
		panic(fmt.Errorf("expected option: '-%s'", opt))
	}
	return
}

func (ca CommandArgs) Get0(opt string) string {
	vs := ca.Get(opt)
	if len(vs) == 0 {
		panic(fmt.Errorf("option '-%s' has no corresponding value", opt))
	}
	return vs[0]
}
func (ca CommandArgs) Get0Default(opt, dft string) string {
	vs, ok := ca[opt]
	if !ok {
		return dft
	}
	if len(vs) == 0 {
		return dft
	}
	return vs[0]
}
func (ca CommandArgs) GetDefault(opt string, i int, dft string) string {
	vs, ok := ca[opt]
	if !ok {
		return dft
	}
	if len(vs) <= i {
		return dft
	}
	return vs[i]
}

func (ca CommandArgs) ContainsOpt(opt string) (ok bool) {
	_, ok = ca[opt]
	return
}

func ParseCommandArgs(args []string) CommandArgs {
	result := make(CommandArgs)
	var i int
	for i < len(args) {
		if isOpt(args[i]) {
			opt := getOpt(args[i])
			i++
			parseVal(opt, args, &i, result)
		} else {
			parseVal("", args, &i, result)
		}
	}
	return result
}

func parseVal(opt string, args []string, i *int, result CommandArgs) {
	var vals []string
	for *i < len(args) && !isOpt(args[*i]) {
		vals = append(vals, args[*i])
		*i++
	}
	result[opt] = vals
}
