struct PropertyBody {
    get: $<><Object>,
    set: Maybe<$<Object><Void>> = Nil
}

struct Property {
    name: String,
    fields of PropertyBody
}


class PropertyCollection is Getter<String, Object>, Setter<String, Object> {
    init (properties: Map<String, PropertyBody>) {
        do nothing
    }
    create (property_list: List) {
        let map = Map<String, PropertyBody>()
        for item in property_list {
            ensure type_correct { item is Property }
            ensure no_conflict(item.name) { not map->has(item.name) }
            set map[item.name] = copy PropertyBody(item)
        }
        return PropertyCollection(map)
        @handle error {
            unless type_correct {
                panic MSG.property_invalid()
            }
            unless no_conflict (name) {
                panic MSG.property_conflict(name)
            }
        }
    }
    get (name: String, loose: Bool) -> Maybe<Object> {
        if not loose {
            ensure property_exists(name) { properties->has(name) }
            return properties[name].get()
        } else {
            if properties->has(name) {
                return properties[name].get()
            } else {
                return Nil
            }
        }
        @handle error {
            unless property_exists (name) {
                panic MSG.property_not_exist(name)
            }
        }
    }
    set (name: String, value: Object) -> Void {
        ensure exists { properties->has(name) }
        ensure writable { properties[name].set is not Nil }
        properties[name].set(value)
        @handle error {
            unless exists {
                panic MSG.property_not_exist(name)
            }
            unless writable {
                panic MSG.property_read_only(name)
            }
        }
    }
    enum_properties () -> Array<Property> {
        return Array<Property>(iterator {
            for { name: body } in properties {
                yield Property { ::name, get: body.get, set: body.set }
            }
        })
    }
}
