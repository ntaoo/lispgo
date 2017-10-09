package printer

import (
	"fmt"
	"strings"
)

import (
	"github.com/ntaoo/lispgo/types"
)

func PrintList(lst []types.Top, printReadable bool,
	start string, end string, join string) string {
	strList := make([]string, 0, len(lst))
	for _, e := range lst {
		strList = append(strList, PrintString(e, printReadable))
	}
	return start + strings.Join(strList, join) + end
}

func PrintString(obj types.Top, printReadable bool) string {
	switch tobj := obj.(type) {
	case types.List:
		return PrintList(tobj.Val, printReadable, "(", ")", " ")
	case types.Vector:
		return PrintList(tobj.Val, printReadable, "[", "]", " ")
	case types.HashMap:
		str_list := make([]string, 0, len(tobj.Val)*2)
		for k, v := range tobj.Val {
			str_list = append(str_list, PrintString(k, printReadable))
			str_list = append(str_list, PrintString(v, printReadable))
		}
		return "{" + strings.Join(str_list, " ") + "}"
	case string:
		if strings.HasPrefix(tobj, "\u029e") {
			return ":" + tobj[2:len(tobj)]
		} else if printReadable {
			return `"` + strings.Replace(
				strings.Replace(
					strings.Replace(tobj, `\`, `\\`, -1),
					`"`, `\"`, -1),
				"\n", `\n`, -1) + `"`
		} else {
			return tobj
		}
	case types.Symbol:
		return tobj.Val
	case nil:
		return "nil"
	case types.MalFunc:
		return "(lambda " +
			PrintString(tobj.Params, true) + " " +
			PrintString(tobj.Exp, true) + ")"
	case func([]types.Top) (types.Top, error):
		return fmt.Sprintf("<function %v>", obj)
	case *types.Atom:
		return "(atom " +
			PrintString(tobj.Val, true) + ")"
	default:
		return fmt.Sprintf("%v", obj)
	}
}
