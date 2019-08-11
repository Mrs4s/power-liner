package power_liner

import (
	"fmt"
	"github.com/mattn/go-runewidth"
	"github.com/peterh/liner"
	"os"
	"reflect"
	"strings"
)

type (
	Shell struct {
		HistoryFile string

		commands   []*Command
		liner      *liner.State
		prompt     string
		onAbort    func(int32)
		abortCount int32
	}

	Command struct {
		Name      string
		Explain   string
		Alias     []string
		Before    func(*Context)
		After     func(*Context)
		Handler   func(*Context)
		Completer func(*Context) []string
	}

	Context struct {
		RawArgs []string
		Nokeys  []string
		Keys    map[string]string
	}

	IHandler interface {
		Alias() map[string][]string
		Explains() map[string]string
	}
)

func NewApp() *Shell {
	shell := &Shell{}
	shell.AppendCommand(&Command{
		Name:    "help",
		Explain: "show help",
		Alias:   []string{"h"},
		Before:  nil,
		After:   nil,
		Handler: func(context *Context) {
			table := [][]string{
				{"Command", "Alias", "Explain"},
			}
			for _, cmd := range shell.commands {
				table = append(table, []string{cmd.Name, strings.Join(cmd.Alias, " "), cmd.Explain})
			}
			shell.PrintTables(table, 2)
		},
		Completer: nil,
	})
	shell.OnAbort(func(i int32) {
		if i < 2 {
			fmt.Println("press Ctrl-C again to exit")
			return
		}
		os.Exit(0)
	})
	return shell
}

func (s *Shell) AddHandler(handler IHandler) {
	t := reflect.TypeOf(handler)
	v := reflect.ValueOf(handler)
	aliasMap := handler.Alias()
	exps := handler.Explains()
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		f, ok := v.Method(i).Interface().(func(*Context))
		if ok {
			var comp func(*Context) []string
			temp := v.MethodByName(m.Name + "Completer")
			if temp.IsValid() {
				comp = temp.Interface().(func(*Context) []string)
			}
			s.AppendCommand(&Command{
				Name:      m.Name,
				Explain:   exps[m.Name],
				Alias:     aliasMap[m.Name],
				Before:    nil,
				After:     nil,
				Handler:   f,
				Completer: comp,
			})
		}
	}
}

func (s *Shell) AppendCommand(cmd *Command) {
	s.commands = append(s.commands, cmd)
}

func (s *Shell) RemoveCommand(name string) {

}

func (s *Shell) SetPrompt(prompt string) {
	s.prompt = prompt
}

func (s *Shell) RunAsShell() {
	s.liner = liner.NewLiner()
	defer s.liner.Close()
	s.liner.SetCtrlCAborts(true)
	s.liner.SetCompleter(func(line string) []string {
		args := parseRaw(line)
		lens := len(filterStrings(args, func(s string) bool { return s != "" }))
		if lens > 0 {
			if cmd := s.filterCommandByNameOrAlias(args[0]); cmd != nil {
				if lens == 1 && !strings.HasSuffix(line, " ") {
					line += " "
				}
				if cmd.Completer != nil {
					pas := cmd.Completer(parse(args))
					if strings.HasSuffix(line, " ") {
						return selectStrings(pas, func(s string) string {
							str := strings.Join(args[:len(args)-1], " ") + " " + s
							if len(args[:len(args)-1]) == 0 {
								str = args[0] + str
							}
							return str
						})
					}
					return selectStrings(filterStrings(pas, func(s string) bool {
						return strings.HasPrefix(strings.ReplaceAll(s, "\"", ""), lastString(args, func(s string) bool {
							return s != ""
						}))
					}), func(s string) string {
						return strings.Join(args[:len(args)-1], " ") + " " + s
					})
				}
			}
		}
		if len(args) <= 1 {
			var cmds []string
			for _, cmd := range s.commands {
				cmds = append(cmds, cmd.Name)
				cmds = append(cmds, cmd.Alias...)
			}
			return filterStrings(cmds, func(s string) bool {
				return strings.HasPrefix(s, line)
			})
		}
		return []string{}
	})
	if file, err := os.Open(s.HistoryFile); err == nil {
		_, _ = s.liner.ReadHistory(file)
		_ = file.Close()
	}
	for {
		if line, err := s.liner.Prompt(s.prompt); err == nil {
			s.liner.AppendHistory(line)
			s.abortCount = 0
			rawArgs := parseRaw(line)
			handleFunc := func(cmd *Command, ctx *Context) {
				defer func() {
					if pan := recover(); pan != nil {
						fmt.Println("error to handle command", cmd.Name, ":", pan)
					}
				}()
				if cmd.Before != nil {
					cmd.Before(ctx)
				}
				if cmd.Handler != nil {
					cmd.Handler(ctx)
				}
				if cmd.After != nil {
					cmd.After(ctx)
				}
			}
			if len(rawArgs) > 0 {
				if command := s.filterCommandByNameOrAlias(rawArgs[0]); command != nil {
					ctx := parse(rawArgs)
					handleFunc(command, ctx)
					continue
				}
			}
			fmt.Println("Unknown command")
		} else if err == liner.ErrPromptAborted {
			s.abortCount++
			if s.onAbort != nil {
				if file, err := os.Open(s.HistoryFile); err == nil {
					_, _ = s.liner.WriteHistory(file)
					_ = file.Close()
				}
				s.onAbort(s.abortCount)
			}
		}
	}
}

func (s *Shell) ReadPassword(prompt string) (string, error) {
	return s.liner.PasswordPrompt(prompt)
}

func (s *Shell) ReadLine(prompt string) (string, error) {
	return s.liner.Prompt(prompt)
}

func (Shell) PrintTables(table [][]string, margin int) {
	var maxLens []int
	for i, col := range table {
		for j, row := range col {
			if i == 0 {
				maxLens = append(maxLens, runewidth.StringWidth(row))
				continue
			}
			length := runewidth.StringWidth(row)
			if maxLens[j] < length {
				maxLens[j] = length
			}
		}
	}
	for _, col := range table {
		for i, row := range col {
			fmt.Print(row)
			if i != len(col)-1 {
				fmt.Print(strings.Repeat(" ", maxLens[i]-runewidth.StringWidth(row)+margin))
			}
		}
		fmt.Println()
	}
}

func (Shell) PrintColumns(strs []string, margin int) {
	maxLength := 0
	marginStr := strings.Repeat(" ", margin)
	var lengths []int
	for _, str := range strs {
		length := runewidth.StringWidth(str)
		maxLength = max(maxLength, length)
		lengths = append(lengths, length)
	}
	width, _, _ := GetTermSize()
	width = int(float32(width) * 1.1)
	numCols, numRows := calculateTableSize(width, margin, maxLength, len(strs))
	if numCols == 1 {
		for _, str := range strs {
			fmt.Println(str)
		}
		return
	}
	for i := 0; i < numCols*numRows; i++ {
		x, y := rowIndexToTableCoords(i, numCols)
		j := tableCoordsToColIndex(x, y, numRows)
		strLen := 0
		str := ""
		if j < len(lengths) {
			strLen = lengths[j]
			str = (strs)[j]
		}
		numSpacesRequired := maxLength - strLen
		spaceStr := strings.Repeat(" ", numSpacesRequired)
		fmt.Print(str)
		if x+1 == numCols {
			fmt.Println()
		} else {
			fmt.Print(spaceStr)
			fmt.Print(marginStr)
		}
	}
}

func (Shell) ClearScreen() {
	_ = ClearScreen()
}

func (s *Shell) OnAbort(f func(int32)) {
	s.onAbort = f
}

func (s *Shell) filterCommandByNameOrAlias(text string) *Command {
	for _, cmd := range s.commands {
		if strings.ToLower(cmd.Name) == strings.ToLower(text) {
			return cmd
		}
		for _, ali := range cmd.Alias {
			if strings.ToLower(ali) == strings.ToLower(text) {
				return cmd
			}
		}
	}
	return nil
}
