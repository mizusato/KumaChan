package syntax

import "strings"


const MAX_NUM_PARTS = 20

type Id int

type Token struct {
    Name     string
    Pattern  Regexp
    Keyword  bool
}

type Rule struct {
    Id        Id
    Nullable  bool
    Branches  [] Branch
}
type Branch struct {
    Parts  [] Part
}
type Part struct {
    Id        Id
    PartType  PartType
    Required  bool
}
type PartType int
const (
    MatchKeyword  PartType  =  iota
    MatchToken
    Recursive
)

func GetPartType (name string) PartType {
    var is_keyword = strings.HasPrefix(name, "@") && len(name) > 1
    if is_keyword {
        return MatchKeyword
    } else {
        var t = name[0:1]
        if strings.ToUpper(t) == t {
            // the name starts with capital letter
            return MatchToken
        } else {
            // the name starts with small letter
            return Recursive
        }
    }
}

func EscapePartName (name string) string {
    if strings.HasPrefix(name, "_") && __EscapeMap[name] != "" {
        return __EscapeMap[name]
    } else {
        return name
    }
}


var __Id2Name = make([] string, 0, 1000)
func Id2Name(id Id) string {
    return __Id2Name[id]
}
var __Name2Id = make(map[string] Id)
func Name2IdMustExist(str string) Id {
    var id, ok = __Name2Id[str]
    if !(ok) { panic("something went wrong") }
    return id
}
func Name2Id(str string) (Id, bool) {
    var id, ok = __Name2Id[str]
    return id, ok
}
var __Id2ConditionalKeyword = make(map[Id] ([] rune))
func Id2ConditionalKeyword(id Id) ([] rune) {
    return __Id2ConditionalKeyword[id]
}
var __Rules = make(map[Id] Rule)
func Rules(id Id) Rule {
    return __Rules[id]
}

func __AssignId2Name (name string) Id {
    var existing, exists = __Name2Id[name]
    if exists {
        return existing
    }
    var id = Id(len(__Id2Name))
    __Name2Id[name] = id
    __Id2Name = append(__Id2Name, name)
    return id
}

func __AssignId2Tokens () {
    for _, token := range __Tokens {
        __AssignId2Name(token.Name)
    }
}

func __AssignId2Keywords () {
    for _, name := range __ConditionalKeywords {
        var keyword = []rune(strings.TrimLeft(name, "@"))
        if len(keyword) == 0 { panic("empty keyword") }
        var id = __AssignId2Name(name)
        __Id2ConditionalKeyword[id] = keyword
    }
}

func __AssignId2Rules () {
    for _, def := range __SyntaxDefinition {
        var t = strings.Split(def, "=")
        var u = strings.Trim(t[0], " ")
        var rule_name = strings.TrimRight(u, "?")
        __AssignId2Name(rule_name)
    }
}

func __ParseRules () {
    for _, def := range __SyntaxDefinition {
        var pivot = strings.Index(def, "=")
        if (pivot == -1) { panic(def + ": invalid rule: missing =") }
        // name = ...
        var str_name = strings.Trim(def[:pivot], " ")
        var name = strings.TrimRight(str_name, "?")
        var nullable = strings.HasSuffix(str_name, "?")
        var id, exists = __Name2Id[name]
        if (!exists) { panic("undefined rule name: " + name) }
        // ... = branches
        var str_branches = strings.Trim(def[pivot+1:], " ")
        if (str_branches == "") { panic(name + ": missing rule definition") }
        var strlist_branches = strings.Split(str_branches, " | ")
        var n_branches = len(strlist_branches)
        var strlist2_branches = make([][]string, n_branches)
        for i, str_branch := range strlist_branches {
            strlist2_branches[i] = strings.Split(str_branch, " ")
        }
        var branches = make([]Branch, n_branches)
        for i, strlist_branch := range strlist2_branches {
            var num_parts = len(strlist_branch)
            branches[i].Parts = make([]Part, num_parts)
            if num_parts > MAX_NUM_PARTS {
                panic(name + ": too many parts")
            }
            for j, str_part := range strlist_branch {
                // check if valid
                if str_part == "" {
                    panic("redundant blank in definition of " + str_name)
                }
                // extract part name
                var required = strings.HasSuffix(str_part, "!")
                var part_name = strings.TrimRight(str_part, "!")
                part_name = EscapePartName(part_name)
                // add to list if it is a keyword
                var part_type = GetPartType(part_name)
                var id, exists = __Name2Id[part_name]
                if (!exists) { panic("undefined part: " + part_name) }
                branches[i].Parts[j] = Part {
                    Id: id, Required: required, PartType: part_type,
                }
            }
        }
        __Rules[id] = Rule {
            Branches: branches,
            Nullable: nullable,
            Id:       id,
        }
    }
}

func __Init () interface{} {
    __AssignId2Tokens()
    __AssignId2Keywords()
    __AssignId2Rules()
    __ParseRules()
    return nil
}

var _ = __Init()
