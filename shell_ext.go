package power_liner

import "math"

func calculateTableSize(width, margin, maxLength, numCells int) (int, int) {
	numCols := (width + margin) / (maxLength + margin)
	if numCols == 0 {
		numCols = 1
	}
	numRows := int(math.Ceil(float64(numCells) / float64(numCols)))
	return numCols, numRows
}

func rowIndexToTableCoords(i, numCols int) (int, int) {
	x := i % numCols
	y := i / numCols
	return x, y
}

func tableCoordsToColIndex(x, y, numRows int) int {
	return y + numRows*x
}

func filterStrings(arr []string, filter func(string) bool) []string {
	var res []string
	for _, str := range arr {
		if filter(str) {
			res = append(res, str)
		}
	}
	return res
}

func lastString(arr []string, filter func(string) bool) string {
	t := filterStrings(arr, filter)
	if len(t) == 0 {
		return ""
	}
	return t[len(t)-1]
}

func selectStrings(arr []string, selector func(string) string) []string {
	var res []string
	for _, str := range arr {
		res = append(res, selector(str))
	}
	return res
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
