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
    array_out_of_range: .{
        'index out of range (index=${1}, length=${2})' << [.index, .length]
    },
    array_inconsistent: .{
        'requested array element (at index ${}) became invalid type' << .index
    },
    array_invalid_slice: .{
        'invalid slice: lower bound is bigger than higher bound'
    },
    array_slice_out_of_range: .{
        'slice index out of range (hi=${1}, length=${2})' << [.hi, .index]
    },
    future_invalid_value: .{
        'resolved future value violated its type constraint'
    },
    stream_invalid_value: .{
        'some element being yielded by the stream violated its type constraint'
    }
}
