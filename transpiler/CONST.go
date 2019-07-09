package transpiler


const (
    RUNTIME = "KumaChan"
    R_VOID = "Void"
    R_EVAL_SCOPE = "Eval"
    R_NEW_SCOPE = "new_scope"
    R_GET_HELPERS = "get_helpers"
    R_REG_MODULE = "register_module"
)


const (
    SCOPE = "scope"
    ERROR_DUMP = "_e"
    DUMP_TYPE = "type"
    DUMP_NAME = "name"
    DUMP_ARGS = "args"
    DUMP_ENSURE = "ee"
    DUMP_TRY = "et"
    TRY_ERROR = "_t_err"
    H_HOOK_ERROR = "_h_err"
    H_HOOK_SCOPE = "_h_scope"
)


const (
    L_METHOD_CALL = "m"
    L_STATIC_SCOPE = "gs"
    L_WRAP = "w"
    L_VAR_LOOKUP = "id"
    L_VAR_DECL = "dl"
    L_VAR_RESET = "rt"
    L_ADD_FUN = "df"
    L_OP_MOUNT = "mnt"
    L_IMPORT_VAR = "ins"
    L_IMPORT_MOD = "im"
    L_IMPORT_ALL = "ia"
    L_GLOBAL_HELPERS = "__"
)


const (
    CALL = "c"
    OPERATOR = "o"
    FOR_LOOP_ITER = "fi"
    FOR_LOOP_ENUM = "fe"
    FOR_LOOP_VALUE = "fv"
    FOR_LOOP_ASYNC = "fa"
    GET = "g"
    SET = "s"
    SLICE = "sl"
    ITER_COMP = "ic"
    LIST_COMP = "lc"
    C_SINGLETON = "cv"
    C_CLASS = "cc"
    C_INTERFACE = "ci"
    C_SCHEMA = "cs"
    C_STRUCT = "ns"
    C_TYPE = "ct"
    C_TEMPLATE = "ctt"
    C_FINITE = "cft"
    C_ENUM = "ce"
    C_FUN_SIG = "cfs"
    REQ_BOOL = "rb"
    REQ_PROMISE = "rp"
    WHEN_FAILED = "wf"
    INJECT_E_ARGS = "ie"
    ENSURE_FAILED = "ef"
    TRY_FAILED = "tf"
    ENTER_H_HOOK = "enh"
    EXIT_H_HOOK = "exh"
    PANIC = "pa"
    ASSERT = "as"
    THROW = "throw"
    T_ANY = "a"
    T_BOOL = "b"
    T_VOID = "v"
    T_TYPE = "t"
    T_HASH = "h"
    T_PROMISE = "pm"
    T_INSTANCE = "i"
    T_ITERATOR = "it"
    T_ASYNC_ITERATOR = "ait"
    T_SLICE_INDEX_DEF = "sid"
)


func G (x string) string {
    return L_GLOBAL_HELPERS + "." + x
}
