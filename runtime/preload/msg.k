let MSG = {
    array_invalid_init: .{
        'invalid array: invalid element was found at index ${}' << .index
    },
    array_out_of_range: .{
        'index out of range (index=${1}, length=${2})' << [.index, .length]
    },
    array_inconsistent: .{
        'requested array element (at index ${}) became invalid type' << .index
    }
}
