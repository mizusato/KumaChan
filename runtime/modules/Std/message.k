let MSG = {
    map_value_not_found: .{
        'requested key does not exist on this map'
    },
    map_inconsistent: .{
        'requested map value became invalid type'
    },
    map_invalid_remove: .{
        'unable to remove entry from map: given key does not exist'
    },
    set_invalid_remove: .{
        'unable to remove element from set: given element does not exist'
    },
    array_invalid_init: .{
        'invalid array: invalid element was found at index ${}' << .index
    },
    array_invalid_index: .{
        'index out of range (index=${1}, length=${2})' << [.index, .length]
    },
    array_inconsistent: .{
        'requested array element (at index ${}) is of invalid type' << .index
    },
    array_invalid_slice: .{
        'invalid slice: lower bound is bigger than higher bound'
    },
    array_invalid_slice_index: .{
        'slice index out of range (high=${1}, length=${2})' << [.high, .index]
    },
    array_slice_inconsistent: .{
        'inconsistent slice: length of underlying array has changed'
    },
    array_unsupported: .{
        '${} is not supported on this array' << .operation
    },
    array_equal_type_invalid: .{
        'cannot check array equality by an equality-undefined element type'
    },
    array_cast_invalid: .{
        'bad array cast: cannot cast an array to a non-array type'
    },
    future_invalid_value: .{
        'resolved future value violated its type constraint'
    },
    stream_invalid_value: .{
        'some element being yielded by the stream violated its type constraint'
    },
    sort_compare_invalid: .{
        'non-boolean value returned by given compare() function'
    }
}
