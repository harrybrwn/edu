// Code generated by "stringer -type School -linecomment school.go"; DO NOT EDIT.

package school

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[UCBerkeley-0]
	_ = x[UCMerced-1]
}

const _School_name = "UCBerkeleyUCMerced"

var _School_index = [...]uint8{0, 10, 18}

func (i School) String() string {
	if i < 0 || i >= School(len(_School_index)-1) {
		return "School(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _School_name[_School_index[i]:_School_index[i+1]]
}