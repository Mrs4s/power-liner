package power_liner

import "strings"

func parseRaw(line string) (rawArgs []string) {
	args := strings.Split(line, " ")
	beginIndex := -1
	endIndex := -1
	for i, arg := range args {
		if strings.HasPrefix(arg, "\"") {
			beginIndex = i
		}
		if strings.HasSuffix(arg, "\"") {
			endIndex = i
		}
		if beginIndex != -1 && endIndex != -1 {
			t := strings.Join(args[beginIndex:endIndex+1], " ")
			rawArgs = append(rawArgs[0:len(rawArgs)-endIndex+beginIndex], t[1:len(t)-1])
			beginIndex = -1
			endIndex = -1
			continue
		}
		rawArgs = append(rawArgs, arg)
	}
	return
}

func parse(line string) *Context {
	args := parseRaw(line)
	allTheArgs := &Context{
		RawArgs: args,
		RawLine: line,
	}
	var nokeyArgsIndices []int
	flags := make(map[int]string)
	allTheArgs.Keys = make(map[string]string)
	for i, arg := range args {
		if len(arg) >= 2 && arg[0:2] == "--" {
			flags[i] = arg[2:]
		} else if len(arg) >= 1 && arg[0:1] == "-" {
			flags[i] = arg[1:]
		} else {
			nokeyArgsIndices = append(nokeyArgsIndices, i)
		}
	}

	for indexOfKey, key := range flags {
		if indexOfKey+1 < len(args) {
			allTheArgs.Keys[key] = checkAndReturnValue(args, indexOfKey)
		} else {
			allTheArgs.Keys[key] = ""
		}
	}

	for _, ind := range nokeyArgsIndices {
		if ind > 0 {
			if getValidValue(args[ind-1]) != "" {
				allTheArgs.Nokeys = append(allTheArgs.Nokeys, args[ind])
			}
		}
	}

	return allTheArgs
}

func checkAndReturnValue(rawArgs []string, keyIndex int) string {
	return getValidValue(rawArgs[keyIndex+1])
}

func getValidValue(val string) string {
	if len(val) >= 2 {
		if val[0:2] == "--" || val[0:1] == "-" {
			return ""
		}
	}
	return val
}
