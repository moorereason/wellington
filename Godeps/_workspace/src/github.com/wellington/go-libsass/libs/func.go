package libs

// #include <stdlib.h>
// #include "sass/context.h"
//
// extern union Sass_Value* GoBridge( union Sass_Value* s_args, void* cookie);
//
// union Sass_Value* CallSassFunction( union Sass_Value* s_args, Sass_Function_Entry cb, struct Sass_Options* opts ) {
//     void* cookie = sass_function_get_cookie(cb);
//     union Sass_Value* ret;
//     ret = GoBridge(s_args, cookie);
//     return ret;
// }
//
import "C"
import "unsafe"

type SassFunc C.Sass_Function_Entry

// SassMakeFunction binds a Go pointer to a Sass function signature
func SassMakeFunction(signature string, ptr unsafe.Pointer) SassFunc {
	csign := C.CString(signature)
	fn := C.sass_make_function(
		csign,
		C.Sass_Function_Fn(C.CallSassFunction),
		ptr)

	return (SassFunc)(fn)
}

var globalFuncs SafeMap

func init() {
	globalFuncs.init()
}

// BindFuncs attaches a slice of Functions to a sass options. Signatures
// are already defined in the SassFunc.
func BindFuncs(opts SassOptions, cookies []Cookie) []*string {

	funcs := make([]SassFunc, len(cookies))
	ids := make([]*string, len(cookies))
	for i, cookie := range cookies {
		idx := globalFuncs.set(cookies[i])

		fn := SassMakeFunction(cookie.Sign,
			unsafe.Pointer(idx))
		funcs[i] = fn
		ids[i] = idx
	}

	sz := C.size_t(len(funcs))
	cfuncs := C.sass_make_function_list(sz)
	for i, cfn := range funcs {
		C.sass_function_set_list_entry(cfuncs, C.size_t(i), cfn)
	}
	C.sass_option_set_c_functions(opts, cfuncs)
	return ids
}

func RemoveFuncs(ids []*string) error {
	for _, idx := range ids {
		delete(globalFuncs.m, idx)
	}
	return nil
}
