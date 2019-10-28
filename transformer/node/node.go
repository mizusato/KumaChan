package node

import (
    "fmt"
    "kumachan/parser/scanner"
    "kumachan/parser/syntax"
    "reflect"
    "strings"
)

type Node struct {
    Point   scanner.Point
    Span    scanner.Span
}

type NodeInfo struct {
    Type       reflect.Type
    Children   map[syntax.Id] NodeChildInfo
    Strings    map[syntax.Id] NodeChildInfo
    Lists      map[syntax.Id] NodeListInfo
    Fallback   map[syntax.Id] NodeFallbackInfo
}

type NodeFallbackInfo struct {
    FieldIndex  int
}

type NodeChildInfo struct {
    FieldIndex  int
    DivePath    [] syntax.Id
    Optional    bool
}

type NodeListInfo struct {
    NodeChildInfo
    ItemId  syntax.Id
    TailId  syntax.Id
}

var __NodeRegistry = []interface{} {
    // Common
    Identifier {},
    // Module
    Module {},
    Resolve {},
    ResolveDetail {},
    Import {},
    ImportedName {},
    // Expressions
    StringLiteral {},
}

var __NodeInfoMap = map[syntax.Id] NodeInfo {}

var __Initialized = false

func __Initialize() {
    var get_field_tag = func(f reflect.StructField) (string, string) {
        var kinds = []string {
            "first", "last",
            "part", "part_opt", "content",
            "list", "list_more", "list_rec",
        }
        for _, kind := range kinds {
            var value = f.Tag.Get(kind)
            if value != "" {
                return kind, value
            }
        }
        return "", ""
    }
    var get_part_id = func(part string) syntax.Id {
        var part_id, exists = syntax.Name2Id[part]
        if !exists {
            panic(fmt.Sprintf("syntax part `%v` does not exist", part))
        }
        return part_id
    }
    var get_parts_id = func(parts []string) []syntax.Id {
        var mapped = make([]syntax.Id, len(parts))
        for i, part := range parts {
            mapped[i] = get_part_id(part)
        }
        return mapped
    }
    var get_dive_info = func(tag_value string) (syntax.Id, []syntax.Id) {
        var t = strings.Split(tag_value, ".")
        if len(t) == 1 {
            return get_part_id(tag_value), []syntax.Id{}
        } else {
            return get_part_id(t[0]), get_parts_id(t[1:])
        }
    }
    for _, node := range __NodeRegistry {
        var T = reflect.TypeOf(node)
        if T.Kind() != reflect.Struct {
            panic("invalid node")
        }
        var f_node, exists = T.FieldByName("Node")
        if !exists {
            panic("invalid node")
        }
        var node_part_name = f_node.Tag.Get("part")
        var node_id = get_part_id(node_part_name)
        var info = NodeInfo {
            Type:     T,
            Children: make(map[syntax.Id] NodeChildInfo),
            Strings:  make(map[syntax.Id] NodeChildInfo),
            Lists:    make(map[syntax.Id] NodeListInfo),
            Fallback: make(map[syntax.Id] NodeFallbackInfo),
        }
        for i := 0; i < T.NumField(); i += 1 {
            var f = T.Field(i)
            var kind, value = get_field_tag(f)
            var part_id, dive_path = get_dive_info(value)
            switch kind {
            case "part":
                info.Children[part_id] = NodeChildInfo{
                    FieldIndex: i,
                    DivePath:   dive_path,
                }
            case "part_opt":
                info.Children[part_id] = NodeChildInfo {
                    FieldIndex: i,
                    DivePath:   dive_path,
                    Optional:   true,
                }
            case "content":
                var part_id = get_part_id(value)
                info.Strings[part_id] = NodeChildInfo {
                    FieldIndex: i,
                    DivePath:   dive_path,
                }
            case "list":
                var list_name string
                if len(dive_path) > 0 {
                    list_name = syntax.Id2Name[dive_path[len(dive_path)-1]]
                } else {
                    list_name = syntax.Id2Name[part_id]
                }
                var t = strings.TrimSuffix(list_name, "list")
                var item = strings.TrimSuffix(t, "_")
                var tail = list_name + "_tail"
                var item_id = get_part_id(item)
                var tail_id = get_part_id(tail)
                info.Lists[part_id] = NodeListInfo{
                    NodeChildInfo: NodeChildInfo{
                        FieldIndex: i,
                        DivePath:   dive_path,
                    },
                    ItemId: item_id,
                    TailId: tail_id,
                }
            case "list_more":
                var item = f.Tag.Get("item")
                if item == "" {
                    panic("`item` tag should be specified when using list_more")
                }
                var item_id = get_part_id(item)
                var tail = "more_" + item + "s"
                var tail_id = get_part_id(tail)
                info.Lists[part_id] = NodeListInfo{
                    NodeChildInfo: NodeChildInfo{
                        FieldIndex: i,
                        DivePath:   dive_path,
                    },
                    ItemId: item_id,
                    TailId: tail_id,
                }
            case "list_rec":
                var tail_id syntax.Id
                if len(dive_path) > 0 {
                    tail_id = dive_path[len(dive_path)-1]
                } else {
                    tail_id = part_id
                }
                var list_name = syntax.Id2Name[tail_id]
                var item = strings.TrimSuffix(list_name, "s")
                var item_id = get_part_id(item)
                info.Lists[part_id] = NodeListInfo{
                    NodeChildInfo: NodeChildInfo{
                        FieldIndex: i,
                        DivePath:   dive_path,
                    },
                    ItemId: item_id,
                    TailId: tail_id,
                }
            default:
                // no tag found, do nothing
            }
            var fallback = f.Tag.Get("fallback")
            if fallback != "" {
                var fallback_id = get_part_id(fallback)
                info.Fallback[fallback_id] = NodeFallbackInfo {
                    FieldIndex: i,
                }
            }
        }
        __NodeInfoMap[node_id] = info
    }
    __Initialized = true
}

func GetNodeInfo(part string) NodeInfo {
    if !__Initialized {
        __Initialize()
        fmt.Printf("%+v\n", __NodeInfoMap)
    }
    var part_id, exists = syntax.Name2Id[part]
    if !exists {
        panic("part " + part + " does not exist")
    } else {
        var info, exists = __NodeInfoMap[part_id]
        if !exists {
            panic("part " + part + " has not been assigned with a node")
        } else {
            return info
        }
    }
}
