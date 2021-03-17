// Code generated by "stringer -output=type_string.go -type=LinkStatusType -trimprefix LinkStatus"; DO NOT EDIT.

package snmp

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[LinkStatusUp-1]
	_ = x[LinkStatusDown-2]
	_ = x[LinkStatusTesting-3]
	_ = x[LinkStatusUnknown-4]
	_ = x[LinkStatusDormant-5]
	_ = x[LinkStatusNotPresent-6]
	_ = x[LinkStatusLowerLayerDown-7]
}

const _LinkStatusType_name = "UpDownTestingUnknownDormantNotPresentLowerLayerDown"

var _LinkStatusType_index = [...]uint8{0, 2, 6, 13, 20, 27, 37, 51}

func (i LinkStatusType) String() string {
	i -= 1
	if i < 0 || i >= LinkStatusType(len(_LinkStatusType_index)-1) {
		return "LinkStatusType(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _LinkStatusType_name[_LinkStatusType_index[i]:_LinkStatusType_index[i+1]]
}