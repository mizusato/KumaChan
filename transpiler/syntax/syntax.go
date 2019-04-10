package syntax

import "strings"

const MAX_NUM_PARTS = 20

type Id int

type Token struct {
    Name     string
    Pattern  Regexp
}

type LeftRight int
const (
    Left  LeftRight = iota
    Right
)

type Operator struct {
    Match     string
    Name      string
    Priority  int
    Assoc     LeftRight
}

type Rule struct {
    Id        Id
    Emptable  bool
    Branches  []Branch
}

type Branch struct {
    Parts  []Part
}

type PartType int

const (
    MatchKeyword  PartType  =  iota
    MatchToken
    Recursive
)

type Part struct {
    Id        Id
    Partype   PartType
    Required  bool
}

func GetPartType (name string) PartType {
    var is_keyword = strings.HasPrefix(name, "@")
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
    if strings.HasPrefix(name, "_") && EscapeMap[name] != "" {
        return EscapeMap[name]
    } else {
        return name
    }
}


var Id2Name []string
var Name2Id map[string]Id
var Id2Keyword map[Id][]rune
var Id2Operator map[Id]Operator
var Rules map[Id]Rule
var RootName string

func Allocate () {
    Id2Name = make([]string, 0, 1000)
    Name2Id = make(map[string]Id)
    Id2Keyword = make(map[Id][]rune)
    Id2Operator = make(map[Id]Operator)
    Rules = make(map[Id]Rule)
}

func AssignId2Name (name string) Id {
    // TODO: check repeat
    var id = Id(len(Id2Name))
    Name2Id[name] = id
    Id2Name = append(Id2Name, name)
    return id
}

func AssignId2Extra () {
    for _, token_name := range Extra {
        AssignId2Name(token_name)
    }
}

func AssignId2Tokens () {
    for _, token := range Tokens {
        AssignId2Name(token.Name)
    }
}

func AssignId2Keywords () {
    for _, name := range Keywords {
        var keyword = []rune(strings.TrimLeft(name, "@"))
        if len(keyword) == 0 { panic("empty keyword") }
        var id = AssignId2Name(name)
        Id2Keyword[id] = keyword
    }
}

func AssignId2Rules () {
    for i, def := range SyntaxDefinition {
        var t = strings.Split(def, "=")
        var u = strings.Trim(t[0], " ")
        var rule_name = strings.TrimRight(u, "?")
        AssignId2Name(rule_name)
        if (i == 0) {
            RootName = rule_name
        }
    }
}

func ConvertOperatorInfo () {
    for _, operator := range Operators {
        var id, exists = Name2Id[operator.Match]
        if exists {
            Id2Operator[id] = operator
        } else {
            panic("match for operator " + operator.Name + " does not exist")
        }
    }
}

func ParseRules () {
    for _, def := range SyntaxDefinition {
        var pivot = strings.Index(def, "=")
        if (pivot == -1) { panic(def + ": invalid rule: missing =") }
        // name = ...
        var str_name = strings.Trim(def[:pivot], " ")
        var name = strings.TrimRight(str_name, "?")
        var emptable = strings.HasSuffix(str_name, "?")
        var id, exists = Name2Id[name]
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
            if num_parts > MAX_NUM_PARTS { panic(name + ": too many parts") }
            for j, str_part := range strlist_branch {
                // extract part name
                var required = strings.HasSuffix(str_part, "!")
                var part_name = strings.TrimRight(str_part, "!")
                part_name = EscapePartName(part_name)
                // add to list if it is a keyword
                var part_type = GetPartType(part_name)
                var id, exists = Name2Id[part_name]
                if (!exists) { panic("undefined part: " + part_name) }
                branches[i].Parts[j] = Part {
                    Id: id, Required: required, Partype: part_type,
                }
            }
        }
        Rules[id] = Rule {
            Branches: branches, Emptable: emptable, Id: id,
        }
    }
}

func Init () {
    Allocate()
    AssignId2Extra()
    AssignId2Tokens()
    AssignId2Keywords()
    AssignId2Rules()
    ConvertOperatorInfo()
    ParseRules()
}
