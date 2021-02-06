// Code generated by "stringer -type=Flags"; DO NOT EDIT.

package ki

import (
	"errors"
	"strconv"
)

var _ = errors.New("dummy error")

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[IsField-0]
	_ = x[HasKiFields-1]
	_ = x[HasNoKiFields-2]
	_ = x[Updating-3]
	_ = x[OnlySelfUpdate-4]
	_ = x[NodeDeleted-5]
	_ = x[NodeDestroyed-6]
	_ = x[ChildAdded-7]
	_ = x[ChildDeleted-8]
	_ = x[ChildrenDeleted-9]
	_ = x[FieldUpdated-10]
	_ = x[PropUpdated-11]
	_ = x[FlagsN-12]
}

const _Flags_name = "IsFieldHasKiFieldsHasNoKiFieldsUpdatingOnlySelfUpdateNodeDeletedNodeDestroyedChildAddedChildDeletedChildrenDeletedFieldUpdatedPropUpdatedFlagsN"

var _Flags_index = [...]uint8{0, 7, 18, 31, 39, 53, 64, 77, 87, 99, 114, 126, 137, 143}

func (i Flags) String() string {
	if i < 0 || i >= Flags(len(_Flags_index)-1) {
		return "Flags(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Flags_name[_Flags_index[i]:_Flags_index[i+1]]
}

func (i *Flags) FromString(s string) error {
	for j := 0; j < len(_Flags_index)-1; j++ {
		if s == _Flags_name[_Flags_index[j]:_Flags_index[j+1]] {
			*i = Flags(j)
			return nil
		}
	}
	return errors.New("String: " + s + " is not a valid option for type: Flags")
}
