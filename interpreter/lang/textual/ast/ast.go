package ast

import (
    "fmt"
    "reflect"
    "strings"
	"kumachan/interpreter/lang/textual/cst"
	"kumachan/interpreter/lang/textual/scanner"
	"kumachan/interpreter/lang/textual/syntax"
)


type Node struct {
    CST    *cst.Tree
    Point  scanner.Point
    Span   scanner.Span
}

type NodeInfo struct {
    Type       reflect.Type
    Children   map[syntax.Id] ([] NodeChildInfo)
    Strings    map[syntax.Id] ([] NodeChildInfo)
    Lists      map[syntax.Id] ([] NodeListInfo)
    Options    map[syntax.Id] ([] NodeChildInfo)
    First      int
    Last       int
}

type NodeChildInfo struct {
    FieldIndex  int
    DivePath    [] syntax.Id
    Optional    bool
    Fallback    syntax.Id
}

type NodeListInfo struct {
    NodeChildInfo
    ItemId  syntax.Id
    TailId  syntax.Id
}

var __NodeRegistry = []interface{} {
    // Root
	Root {},
	Shebang {},
	ReplRoot {},
	ReplAssign {},
	ReplDo {},
	ReplEval {},
    // Statement
    VariousStatement {},
    Title {},
    Import {},
    Identifier {},
    Do {},
    DeclConst {},
    VariousConstValue {},
    DeclFunction {},
    VariousBody {},
    NativeRef {},
    DeclType {},
    Doc {},
	Tag {},
	VariousTypeDef {},
    NativeType {},
    ImplicitType {},
	BoxedType {},
    EnumType {},
    TypeParam {},
    TypeParamDefault {},
    VariousTypeBound {},
    TypeLowerBound {},
    TypeHigherBound {},
    // Type
    VariousType {},
    TypeRef {},
    TypeLiteral {},
    VariousRepr {},
    ReprTuple {},
    ReprRecord {},
    Field {},
    ReprFunc {},
    // Expression
    Expr {},
    VariousTerm {},
    VariousPipe {},
    PipeFunc {},
    PipeGet {},
	PipeRefField {},
    PipeCast {},
    PipeSwitch {},
    PipeRefBranch {},
    // Term
    VariousCall {},
    CallPrefix {},
    CallInfix {},
	Switch {},
    Branch {},
    MultiSwitch {},
    MultiBranch {},
    If {},
    ElIf {},
    Lambda {},
    PipelineLambda {},
    VariousPattern {},
	PatternTrivial {},
    PatternTuple {},
    PatternRecord {},
    FieldMap {},
	Array {},
    Tuple {},
    Record {},
    FieldValue {},
    Update {},
    Block {},
    Binding {},
    Cps {},
    CpsBinding {},
    InlineRef {},
    IntegerLiteral {},
    FloatLiteral {},
    StringLiteral {},
    VariousStringPart {},
    StringText {},
    Formatter {},
    VariousFormatterPart {},
    FormatterText {},
    CharLiteral {},
}

var __NodeInfoMap = map[syntax.Id] NodeInfo {}

var __Initialized = false

func __Initialize() {
    var get_field_tag = func(f reflect.StructField) (string, string) {
        var kinds = []string {
            "use",
            "part", "part_opt", "content",
            "option",
            "list", "list_more", "list_rec",
        }
        for _, kind := range kinds {
            var value, exists = f.Tag.Lookup(kind)
            if exists {
                return kind, value
            }
        }
        return "", ""
    }
    var get_part_id = func(part string) syntax.Id {
        var part_id, exists = syntax.Name2Id(part)
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
        var path = strings.Split(tag_value, ".")
        if len(path) == 1 && path[0] == "" {
            return (syntax.Id)(-1), []syntax.Id {}
        } else {
            return get_part_id(path[0]), get_parts_id(path)
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
            Children: make(map[syntax.Id] ([] NodeChildInfo)),
            Strings:  make(map[syntax.Id] ([] NodeChildInfo)),
            Lists:    make(map[syntax.Id] ([] NodeListInfo)),
            Options:  make(map[syntax.Id] ([] NodeChildInfo)),
            First:    -1,
            Last:     -1,
        }
        for i := 0; i < T.NumField(); i += 1 {
            var f = T.Field(i)
            if f.Name == "Node" {
                continue
            }
            var kind, value = get_field_tag(f)
            if kind == "use" {
                if value == "first" {
                    info.First = i
                } else if value == "last" {
                    info.Last = i
                } else {
                    panic(fmt.Sprintf("invalid directive `use:'%v'`", value))
                }
                continue
            }
            var part_id, dive_path = get_dive_info(value)
            var fallback = f.Tag.Get("fallback")
            var fallback_id syntax.Id = -1
            if fallback != "" {
                fallback_id = get_part_id(fallback)
            }
            switch kind {
            // case "use": already handled above
            case "part":
                info.Children[part_id] = append(info.Children[part_id], NodeChildInfo {
                    FieldIndex: i,
                    DivePath:   dive_path,
                    Fallback:   fallback_id,
                })
            case "part_opt":
                info.Children[part_id] = append(info.Children[part_id], NodeChildInfo {
                    FieldIndex: i,
                    DivePath:   dive_path,
                    Optional:   true,
                    Fallback:   fallback_id,
                })
            case "content":
                info.Strings[part_id] = append(info.Strings[part_id], NodeChildInfo {
                    FieldIndex: i,
                    DivePath:   dive_path,
                    Fallback:   fallback_id,
                })
            case "option":
                info.Options[part_id] = append(info.Options[part_id], NodeChildInfo {
                    FieldIndex: i,
                    DivePath:   dive_path,
                    Optional:   true,
                })
            case "list":
                var list_name string
                if len(dive_path) > 0 {
                    list_name = syntax.Id2Name(dive_path[len(dive_path)-1])
                } else {
                    list_name = syntax.Id2Name(part_id)
                }
                var t = strings.TrimSuffix(list_name, "list")
                var item = strings.TrimSuffix(t, "_")
                var tail = list_name + "_tail"
                var item_id = get_part_id(item)
                var tail_id = get_part_id(tail)
                info.Lists[part_id] = append(info.Lists[part_id], NodeListInfo {
                    NodeChildInfo: NodeChildInfo {
                        FieldIndex: i,
                        DivePath:   dive_path,
                        Optional:   true,
                        Fallback:   fallback_id,
                    },
                    ItemId: item_id,
                    TailId: tail_id,
                })
            case "list_more":
                var item = f.Tag.Get("item")
                if item == "" {
                    panic("`item` tag should be specified when using list_more")
                }
                var item_id = get_part_id(item)
                var tail string
                if strings.HasSuffix(item, "ch") || strings.HasSuffix(item, "x") {
                    tail = "more_" + item + "es"
                } else {
                    tail = "more_" + item + "s"
                }
                var tail_id = get_part_id(tail)
                info.Lists[part_id] = append(info.Lists[part_id], NodeListInfo {
                    NodeChildInfo: NodeChildInfo {
                        FieldIndex: i,
                        DivePath:   dive_path,
                        Optional:   true,
                        Fallback:   fallback_id,
                    },
                    ItemId: item_id,
                    TailId: tail_id,
                })
            case "list_rec":
                var tail_id syntax.Id
                if len(dive_path) > 0 {
                    tail_id = dive_path[len(dive_path)-1]
                } else {
                    tail_id = part_id
                }
                var list_name = syntax.Id2Name(tail_id)
                var item string
                if strings.HasSuffix(list_name, "ches") {
                    item = strings.TrimSuffix(list_name, "es")
                } else {
                    item = strings.TrimSuffix(list_name, "s")
                }
                var item_id = get_part_id(item)
                info.Lists[part_id] = append(info.Lists[part_id], NodeListInfo {
                    NodeChildInfo: NodeChildInfo {
                        FieldIndex: i,
                        DivePath:   dive_path,
                        Optional:   true,
                        Fallback:   fallback_id,
                    },
                    ItemId: item_id,
                    TailId: tail_id,
                })
            default:
                // no tag found, do nothing
            }
        }
        __NodeInfoMap[node_id] = info
    }
    __Initialized = true
}

func GetTagContent(tag Tag) string {
    var t = string(tag.RawContent)
    t = strings.TrimSuffix(t, "\r")
    t = strings.TrimPrefix(t, "#")
    t = strings.Trim(t, " \t")
    return t
}

func GetNodeInfoById (part_id syntax.Id) NodeInfo {
    if !__Initialized {
        __Initialize()
    }
    var info, exists = __NodeInfoMap[part_id]
    if !exists {
        panic(fmt.Sprintf (
            "node info of part `%v` does not exist",
            syntax.Id2Name(part_id),
        ))
    } else {
        return info
    }
}

