const LIST_INIT_SIZE = 30  // even integer
const LIST_GROW = 10  // grow factor
const LIST_WASTE_BOUND = 8  // ratio to 12 (0 < integer < 12)


func assert_valid_element(x interface{}) {
    if x == nil {
        panic("nil element in linear list")
    }
}


type LinearList struct {
    data []Object
    head int
    tail int
    center int
}


func MakeList() *LinearList {
    return &LinearList {
        data: make([]Object, LIST_INIT_SIZE*2+1),
        head: LIST_INIT_SIZE,
        tail: LIST_INIT_SIZE,
        center: LIST_INIT_SIZE,
    }
}


func (l *LinearList) length() int {
    return (l.tail - l.head)
}


func (l *LinearList) assert_index(n int) {
    if !(0 <= n && n < l.length()) {
        panic("linear list index error")
    }
}


func (l *LinearList) assert_full() {
    if l.length() == 0 {
        panic("invalid operation on empty list")
    }
}


func (l *LinearList) at(index int) Object {
    l.assert_index(index)
    value := l.data[l.head + index]
    assert_valid_element(value)
    return value
}


func (l *LinearList) replace(index int, new_value Object) {
    l.assert_index(index)
    l.data[l.head + index] = new_value
}


func (l *LinearList) first() Object {
    l.assert_full()
    value := l.data[l.head]
    assert_valid_element(value)
    return value
}


func (l *LinearList) last() Object {
    l.assert_full()
    value := l.data[l.tail-1]
    assert_valid_element(value)
    return value
}


func (l *LinearList) prepend(element Object) {
    capacity := len(l.data)
    if (l.head == 0) {
        offset_backward := l.center
        grow_backward := offset_backward * LIST_GROW
        delta := grow_backward - offset_backward
        new_capacity := capacity + delta
        new_data := make([]Object, new_capacity)
        for i := l.head; i < l.tail; i++ {
            new_data[i+delta] = l.data[i]
        }
        l.data = new_data
        l.center += delta
        l.head += delta
        l.tail += delta
        // fmt.Printf("grow: %v -> %v\n", capacity, new_capacity)
    }
    if (l.head != l.tail) {
        l.head -= 1
    }
    l.data[l.head] = element
}


func (l *LinearList) shift() {
    if (l.head != l.tail) {
        l.data[l.head] = nil
        l.head += 1
        l.check_waste()
    } else {
        panic("linear list has no element to shift")
    }
}


func (l *LinearList) append(element Object) {
    capacity := len(l.data)
    if (l.tail == capacity) {
        offset_forward := (capacity - l.center) - 1
        grow_forward := offset_forward * LIST_GROW
        new_capacity := capacity + (grow_forward - offset_forward)
        new_data := make([]Object, new_capacity)
        for i := l.head; i < l.tail; i++ {
            new_data[i] = l.data[i]
        }
        l.data = new_data
        // fmt.Printf("grow: %v -> %v\n", capacity, new_capacity)
    }
    l.data[l.tail] = element
    l.tail += 1
}


func (l *LinearList) pop() {
    if (l.head != l.tail) {
        l.tail -= 1
        l.data[l.tail] = nil
        l.check_waste()
    } else {
        panic("linear list has no element to pop")
    }
}


func (l *LinearList) insert_left(position int, element Object) {
    l.assert_index(position)
    if (l.length() == 0) {
        l.append(element)
    } else {
        var last_element = l.data[l.tail-1]
        for i := l.tail-1; i > l.head + position; i-- {
            l.data[i] = l.data[i-1]
        }
        l.data[l.head + position] = element
        l.append(last_element)
    }
}


func (l *LinearList) insert_right(position int, element Object) {
    l.assert_index(position)
    if (l.length() == 0) {
        l.prepend(element)
    } else {
        var first_element = l.data[l.head]
        for i := l.head; i < l.head + position; i++ {
            l.data[i] = l.data[i+1]
        }
        l.data[l.head + position] = element
        l.prepend(first_element)
    }
}


func (l *LinearList) remove(position int) {
    l.assert_index(position)
    for i := l.head + position; i < l.tail-1; i++ {
        l.data[i] = l.data[i+1]
    }
    l.data[l.tail-1] = nil
    l.tail -= 1
}


func (l *LinearList) check_waste() {
    capacity := len(l.data)
    length := l.tail - l.head
    if length == 0 {
        /* if the last element has been removed, move head to center */
        l.head = l.center
        l.tail = l.head
    } else {
        /* if head or tail ran far away from center, pull data back */
        run_forward := l.head - l.center
        run_backward := l.center - (l.tail-1)
        bound_forward := (capacity - l.center - 1) * LIST_WASTE_BOUND / 12
        bound_backward := (l.center) * LIST_WASTE_BOUND / 12
        if run_forward > bound_forward && run_forward >= length {
            new_head := l.center
            for i := 0; i < length; i++ {
                old_index := l.head + i
                l.data[new_head + i] = l.data[old_index]
                l.data[old_index] = nil
            }
            l.head = new_head
            l.tail = l.head + length
        } else if run_backward > bound_backward && run_backward >= length {
            new_head := (l.center+1) - length
            for i := 0; i < length; i++ {
                old_index := l.head + i
                l.data[new_head + i] = l.data[old_index]
                l.data[old_index] = nil
            }
            l.head = new_head
            l.tail = l.head + length
        }
    }
}

