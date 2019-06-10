let MSG = {
    array_invalid_init: .{
        'invalid array: invalid element was found at index ${}' << .index
    },
    array_out_of_range: .{
        'index out of range (index=${1}, length=${2})' << [.index, .length]
    },
    array_inconsistent: .{
        'requested array element (at index ${}) became invalid type' << .index
    },
    map_value_not_found: .{
        'requested key does not exist on this map'
    },
    map_inconsistent: .{
        'requested map value became invalid type'
    },
    map_invalid_remove: .{
        'cannot remove non-existing element from map'
    },
    future_invalid_value: .{
        'resolved future value violated its type constraint'
    }
}
