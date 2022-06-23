// File generated by docs_parser script. DO NOT EDIT!

package parse

var BuiltinTypes = map[string]*JavaType{
    "String": {
        Name: "String",
        PackageName: "java.lang",
        Visibility: VisibilityPublic,
    },
}

func init() {

    BuiltinTypes["String"].Fields = []*JavaField{
        {
            Name: "CASE_INSENSITIVE_ORDER",
            Visibility: VisibilityPublic,
            Type: getOrCreateBuiltinType("Comparator"),
            IsStatic: true,
            IsFinal: true,
        },
    }

    BuiltinTypes["String"].Methods = []*JavaMethod{
        {
            Name: "chars",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("IntStream"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "codePointAt",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("int"),
            Arguments: []*JavaArgument{
                {
                    Name: "index",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "codePointBefore",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("int"),
            Arguments: []*JavaArgument{
                {
                    Name: "index",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "codePointCount",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("int"),
            Arguments: []*JavaArgument{
                {
                    Name: "beginIndex",
                    Type: getOrCreateBuiltinType("int"),
                },
                {
                    Name: "endIndex",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "codePoints",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("IntStream"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "compareTo",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("int"),
            Arguments: []*JavaArgument{
                {
                    Name: "anotherString",
                    Type: getOrCreateBuiltinType("String"),
                },
            },
        },
        {
            Name: "compareToIgnoreCase",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("int"),
            Arguments: []*JavaArgument{
                {
                    Name: "str",
                    Type: getOrCreateBuiltinType("String"),
                },
            },
        },
        {
            Name: "concat",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "str",
                    Type: getOrCreateBuiltinType("String"),
                },
            },
        },
        {
            Name: "contains",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("boolean"),
            Arguments: []*JavaArgument{
                {
                    Name: "s",
                    Type: getOrCreateBuiltinType("CharSequence"),
                },
            },
        },
        {
            Name: "contentEquals",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("boolean"),
            Arguments: []*JavaArgument{
                {
                    Name: "cs",
                    Type: getOrCreateBuiltinType("CharSequence"),
                },
            },
        },
        {
            Name: "contentEquals",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("boolean"),
            Arguments: []*JavaArgument{
                {
                    Name: "sb",
                    Type: getOrCreateBuiltinType("StringBuffer"),
                },
            },
        },
        {
            Name: "copyValueOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "data",
                    Type: getOrCreateBuiltinType("char[]"),
                },
            },
        },
        {
            Name: "copyValueOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "data",
                    Type: getOrCreateBuiltinType("char[]"),
                },
                {
                    Name: "offset",
                    Type: getOrCreateBuiltinType("int"),
                },
                {
                    Name: "count",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "describeConstable",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("Optional"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "endsWith",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("boolean"),
            Arguments: []*JavaArgument{
                {
                    Name: "suffix",
                    Type: getOrCreateBuiltinType("String"),
                },
            },
        },
        {
            Name: "equals",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("boolean"),
            Arguments: []*JavaArgument{
                {
                    Name: "anObject",
                    Type: getOrCreateBuiltinType("Object"),
                },
            },
        },
        {
            Name: "equalsIgnoreCase",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("boolean"),
            Arguments: []*JavaArgument{
                {
                    Name: "anotherString",
                    Type: getOrCreateBuiltinType("String"),
                },
            },
        },
        {
            Name: "format",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "format",
                    Type: getOrCreateBuiltinType("String"),
                },
                {
                    Name: "args",
                    Type: getOrCreateBuiltinType("Object..."),
                },
            },
        },
        {
            Name: "format",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "l",
                    Type: getOrCreateBuiltinType("Locale"),
                },
                {
                    Name: "format",
                    Type: getOrCreateBuiltinType("String"),
                },
                {
                    Name: "args",
                    Type: getOrCreateBuiltinType("Object..."),
                },
            },
        },
        {
            Name: "formatted",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "args",
                    Type: getOrCreateBuiltinType("Object..."),
                },
            },
        },
        {
            Name: "getBytes",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("byte[]"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "getBytes",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("void"),
            Arguments: []*JavaArgument{
                {
                    Name: "srcBegin",
                    Type: getOrCreateBuiltinType("int"),
                },
                {
                    Name: "srcEnd",
                    Type: getOrCreateBuiltinType("int"),
                },
                {
                    Name: "dst",
                    Type: getOrCreateBuiltinType("byte[]"),
                },
                {
                    Name: "dstBegin",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "getBytes",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("byte[]"),
            Arguments: []*JavaArgument{
                {
                    Name: "charsetName",
                    Type: getOrCreateBuiltinType("String"),
                },
            },
        },
        {
            Name: "getBytes",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("byte[]"),
            Arguments: []*JavaArgument{
                {
                    Name: "charset",
                    Type: getOrCreateBuiltinType("Charset"),
                },
            },
        },
        {
            Name: "getChars",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("void"),
            Arguments: []*JavaArgument{
                {
                    Name: "srcBegin",
                    Type: getOrCreateBuiltinType("int"),
                },
                {
                    Name: "srcEnd",
                    Type: getOrCreateBuiltinType("int"),
                },
                {
                    Name: "dst",
                    Type: getOrCreateBuiltinType("char[]"),
                },
                {
                    Name: "dstBegin",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "hashCode",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("int"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "indent",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "n",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "indexOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("int"),
            Arguments: []*JavaArgument{
                {
                    Name: "ch",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "indexOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("int"),
            Arguments: []*JavaArgument{
                {
                    Name: "ch",
                    Type: getOrCreateBuiltinType("int"),
                },
                {
                    Name: "fromIndex",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "indexOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("int"),
            Arguments: []*JavaArgument{
                {
                    Name: "str",
                    Type: getOrCreateBuiltinType("String"),
                },
            },
        },
        {
            Name: "indexOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("int"),
            Arguments: []*JavaArgument{
                {
                    Name: "str",
                    Type: getOrCreateBuiltinType("String"),
                },
                {
                    Name: "fromIndex",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "intern",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "isBlank",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("boolean"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "isEmpty",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("boolean"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "join",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "delimiter",
                    Type: getOrCreateBuiltinType("CharSequence"),
                },
                {
                    Name: "elements",
                    Type: getOrCreateBuiltinType("CharSequence..."),
                },
            },
        },
        {
            Name: "join",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "delimiter",
                    Type: getOrCreateBuiltinType("CharSequence"),
                },
                {
                    Name: "extends",
                    Type: getOrCreateBuiltinType("Iterable<?"),
                },
            },
        },
        {
            Name: "lastIndexOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("int"),
            Arguments: []*JavaArgument{
                {
                    Name: "ch",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "lastIndexOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("int"),
            Arguments: []*JavaArgument{
                {
                    Name: "ch",
                    Type: getOrCreateBuiltinType("int"),
                },
                {
                    Name: "fromIndex",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "lastIndexOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("int"),
            Arguments: []*JavaArgument{
                {
                    Name: "str",
                    Type: getOrCreateBuiltinType("String"),
                },
            },
        },
        {
            Name: "lastIndexOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("int"),
            Arguments: []*JavaArgument{
                {
                    Name: "str",
                    Type: getOrCreateBuiltinType("String"),
                },
                {
                    Name: "fromIndex",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "length",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("int"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "lines",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("Stream"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "matches",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("boolean"),
            Arguments: []*JavaArgument{
                {
                    Name: "regex",
                    Type: getOrCreateBuiltinType("String"),
                },
            },
        },
        {
            Name: "offsetByCodePoints",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("int"),
            Arguments: []*JavaArgument{
                {
                    Name: "index",
                    Type: getOrCreateBuiltinType("int"),
                },
                {
                    Name: "codePointOffset",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "regionMatches",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("boolean"),
            Arguments: []*JavaArgument{
                {
                    Name: "ignoreCase",
                    Type: getOrCreateBuiltinType("boolean"),
                },
                {
                    Name: "toffset",
                    Type: getOrCreateBuiltinType("int"),
                },
                {
                    Name: "other",
                    Type: getOrCreateBuiltinType("String"),
                },
                {
                    Name: "ooffset",
                    Type: getOrCreateBuiltinType("int"),
                },
                {
                    Name: "len",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "regionMatches",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("boolean"),
            Arguments: []*JavaArgument{
                {
                    Name: "toffset",
                    Type: getOrCreateBuiltinType("int"),
                },
                {
                    Name: "other",
                    Type: getOrCreateBuiltinType("String"),
                },
                {
                    Name: "ooffset",
                    Type: getOrCreateBuiltinType("int"),
                },
                {
                    Name: "len",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "repeat",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "count",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "replace",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "oldChar",
                    Type: getOrCreateBuiltinType("char"),
                },
                {
                    Name: "newChar",
                    Type: getOrCreateBuiltinType("char"),
                },
            },
        },
        {
            Name: "replace",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "target",
                    Type: getOrCreateBuiltinType("CharSequence"),
                },
                {
                    Name: "replacement",
                    Type: getOrCreateBuiltinType("CharSequence"),
                },
            },
        },
        {
            Name: "replaceAll",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "regex",
                    Type: getOrCreateBuiltinType("String"),
                },
                {
                    Name: "replacement",
                    Type: getOrCreateBuiltinType("String"),
                },
            },
        },
        {
            Name: "replaceFirst",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "regex",
                    Type: getOrCreateBuiltinType("String"),
                },
                {
                    Name: "replacement",
                    Type: getOrCreateBuiltinType("String"),
                },
            },
        },
        {
            Name: "resolveConstantDesc",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "lookup",
                    Type: getOrCreateBuiltinType("MethodHandles.Lookup"),
                },
            },
        },
        {
            Name: "split",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String[]"),
            Arguments: []*JavaArgument{
                {
                    Name: "regex",
                    Type: getOrCreateBuiltinType("String"),
                },
            },
        },
        {
            Name: "split",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String[]"),
            Arguments: []*JavaArgument{
                {
                    Name: "regex",
                    Type: getOrCreateBuiltinType("String"),
                },
                {
                    Name: "limit",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "startsWith",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("boolean"),
            Arguments: []*JavaArgument{
                {
                    Name: "prefix",
                    Type: getOrCreateBuiltinType("String"),
                },
            },
        },
        {
            Name: "startsWith",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("boolean"),
            Arguments: []*JavaArgument{
                {
                    Name: "prefix",
                    Type: getOrCreateBuiltinType("String"),
                },
                {
                    Name: "toffset",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "strip",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "stripIndent",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "stripLeading",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "stripTrailing",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "subSequence",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("CharSequence"),
            Arguments: []*JavaArgument{
                {
                    Name: "beginIndex",
                    Type: getOrCreateBuiltinType("int"),
                },
                {
                    Name: "endIndex",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "substring",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "beginIndex",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "substring",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "beginIndex",
                    Type: getOrCreateBuiltinType("int"),
                },
                {
                    Name: "endIndex",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "toCharArray",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("char[]"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "toLowerCase",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "toLowerCase",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "locale",
                    Type: getOrCreateBuiltinType("Locale"),
                },
            },
        },
        {
            Name: "toString",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "toUpperCase",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "toUpperCase",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "locale",
                    Type: getOrCreateBuiltinType("Locale"),
                },
            },
        },
        {
            Name: "transform",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("R"),
            Arguments: []*JavaArgument{
                {
                    Name: "super",
                    Type: getOrCreateBuiltinType("Function<?"),
                },
                {
                    Name: "extends",
                    Type: getOrCreateBuiltinType("?"),
                },
            },
        },
        {
            Name: "translateEscapes",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "trim",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
            },
        },
        {
            Name: "valueOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "b",
                    Type: getOrCreateBuiltinType("boolean"),
                },
            },
        },
        {
            Name: "valueOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "c",
                    Type: getOrCreateBuiltinType("char"),
                },
            },
        },
        {
            Name: "valueOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "data",
                    Type: getOrCreateBuiltinType("char[]"),
                },
            },
        },
        {
            Name: "valueOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "data",
                    Type: getOrCreateBuiltinType("char[]"),
                },
                {
                    Name: "offset",
                    Type: getOrCreateBuiltinType("int"),
                },
                {
                    Name: "count",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "valueOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "d",
                    Type: getOrCreateBuiltinType("double"),
                },
            },
        },
        {
            Name: "valueOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "f",
                    Type: getOrCreateBuiltinType("float"),
                },
            },
        },
        {
            Name: "valueOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "i",
                    Type: getOrCreateBuiltinType("int"),
                },
            },
        },
        {
            Name: "valueOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "l",
                    Type: getOrCreateBuiltinType("long"),
                },
            },
        },
        {
            Name: "valueOf",
            Visibility: VisibilityPublic,
            ReturnType: getOrCreateBuiltinType("String"),
            Arguments: []*JavaArgument{
                {
                    Name: "obj",
                    Type: getOrCreateBuiltinType("Object"),
                },
            },
        },
    }
}
