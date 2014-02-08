//  ---------------------------------------------------------------------------
//
//  main.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

// genproc is a code generation tool for automatically creating pack/unpack
// routines which work with goat's built-in net.Protocol implementation.
// The application recursively searches a path for .go files and parses them, 
// looking for structs which are marked up as network messages which also have
// fields marked for network export. For any relevant structs, a new type with
// the name <structName>Handler will be generated which implements the 
// net.MsgProcessor interface. The new type and associated functions will be 
// saved in a new file msg<structName>.go in the same directory as the source
// file which contained the marked up source struct.
//  Usage: genproc <search root directory>
//  
//  // Mark-up structs which you'll need to serialize/deserialize with a
//  /* +NetMsg+ <message signature> */ tag
//  /* +NetMsg+ 25 */
//  type ExampleNetMsg struct {
//      Field1 int      // +export+
//      Field2 string
//      Field3 string   // +export+
//  }
// In this example, a new type ExampleNetMsgHandler will be created with
// relevant Close, Init, Deserialize, Serialize, and Signature functions.
// The Signature function will return the message signature contained in the
// +NetMsg+ comment tag (in this case 25, though it could also be a variable
// reference). The Serialize an Deserialize methods will pack and unpack 
// Field1 and Field3 using the lib/buffer package. The user can then register 
// ExNetMsgHandler with his/her chosen net.Protocol.
package main

// External imports.
import (
    "github.com/xaevman/goat/lib/fs"
)

// Stdlib imports.
import (
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "text/template"
    "unicode"
    "unicode/utf8"
)

// Field export comment flag.
const FIELD_EXPORT_FLAG = "+export+"

// NetMsg comment regexp.
var netMsgRegexp = regexp.MustCompile("/\\* \\+NetMsg\\+ (.*) \\*/")


// newMsgGenData is a constructor helper which creates a new MsgGenData object
// and returns a pointer to it for use.
func newMsgGenData() *MsgGenData {
    data        := new(MsgGenData)
    data.Exports = make([]MsgGenExport, 0)
    return data
}

// MsgGenData represents code generation metadata about a specific +NetMsg+
// flagged type.
type MsgGenData struct {
    Exports []MsgGenExport
    Imports []string
    MsgSig  string
    NetMsg  string
    Package string
    Path    string
}


// MsgGenExport represents code generation metadata about an exported field
// with a +NetMsg+ flagged type.
type MsgGenExport struct {
    Name string
    Type string
}

// File system search objects.
var (
    searcher   = fs.NewSearchDir()
    searchPath = "."
)


// main is the application's entry point.
func main() {
    // set root search path
    if len(os.Args) > 1 {
        searchPath = os.Args[1]
    }

    fmt.Println("Searching for net.Msg specs...")
    fmt.Println()

    // find and parse .go files
    go searcher.SearchFiles(searchPath, "*.go")

    func() {
        for {
            select {
            case err := <- searcher.ErrChan:
                fmt.Println(err)
            case file := <- searcher.FileChan:
                parseFile(file)
            case <-searcher.DoneChan:
                return
            }
        }
    }()

    // flush any remaining data
    func () {
        for {
            select {
            default: return
            case err := <- searcher.ErrChan:
                fmt.Println(err)
            case file := <- searcher.FileChan:
                parseFile(file)
            }
        }
    }()

    fmt.Println()
}

// parseFile takes a file path and attempts to parse it, looking for marked
// up network messages.
func parseFile(path string) {
    fset := token.NewFileSet()

    // parse imports first
    imports := parseImports(fset, path)

    // parse structs looking for message defs
    parseStructs(fset, path, imports)
}

// parseImports takes a given fileset and path, and tries to pull any import
// statements out. The import statements will be passed along to the struct
// parser for inclusion in the exported code file.
func parseImports(fset *token.FileSet, path string) []string {
    // parse imports first
    f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
    if err != nil {
        panic(err)
    }

    imports := make([]string, 0)

    for _, decl := range f.Decls {
        gDecl, ok := decl.(*ast.GenDecl)
        if !ok {
            continue
        }

        if gDecl.Specs == nil {
            continue
        }

        // Loop through specs looking for import statements
        for _, spec := range gDecl.Specs {
            sDecl, ok := spec.(*ast.ImportSpec)
            if !ok {
                continue
            }

            if sDecl.Path == nil {
                continue
            }

            imports = append(imports, sDecl.Path.Value)
        }
    }

    return imports
}

// parseStructs takes a given fileset, path and list of imports and attempts
// to find structs which have been marked up using the +NetMsg+ flag. If 
// it is able to parse the source file, and a give struct has one or more fields
// marked for export, the template will be applied and a new source file created.
func parseStructs(fset *token.FileSet, path string, imports []string) {
    // Parse the file containing this very example
    // but stop after processing the imports.
    f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
    if err != nil {
        panic(err)
    }

    packName     := f.Name.Name
    parseObjects := make(map[*MsgGenData]*ast.GenDecl, 0)

    // Traverse all comments, looking for declarations flagged for code
    // generation.
    for _, decl := range f.Decls {
        gDecl, ok := decl.(*ast.GenDecl)
        if !ok {
            continue
        }

        if gDecl.Doc == nil ||  gDecl.Doc.List == nil {
            continue
        }

        for _, comment := range gDecl.Doc.List {
            sig, valid := validateStructComment(comment.Text)
            if valid {
                newObj        := newMsgGenData()
                newObj.Imports = imports
                newObj.MsgSig  = sig
                newObj.Package = packName
                newObj.Path    = filepath.Dir(path)

                parseObjects[newObj] = gDecl

                fmt.Printf("path: %s\n", path)
                fmt.Printf("package %s, sig %s added\n", packName, sig)
            }
        }
    }

    // loop over decls to create metadata objects
    for obj, decl := range parseObjects {
    for _, spec   := range decl.Specs {
        tSpec, ok := spec.(*ast.TypeSpec)
        if !ok {
            continue
        }

        obj.NetMsg = tSpec.Name.Name
        fmt.Printf("\tstruct name: %s\n", obj.NetMsg)

        sType, ok := tSpec.Type.(*ast.StructType)
        if !ok {
            continue
        }

        for _, field   := range sType.Fields.List {
            if field.Comment == nil {
                continue
            }

            for _, comment := range field.Comment.List {
                if !strings.Contains(comment.Text, FIELD_EXPORT_FLAG) {
                    continue
                }

                // exported field
                if field.Names == nil || len(field.Names) != 1 {
                    continue
                }

                if field.Type == nil {
                    continue
                }

                tDecl, ok := field.Type.(*ast.Ident)
                if !ok {
                    continue
                }

                export := MsgGenExport { field.Names[0].Name, setCaps(tDecl.Name) }
                obj.Exports = append(obj.Exports, export)

                fmt.Printf("\t\texport %s :: %s\n", export.Name, export.Type)
            }
        }

        if  !validateMsgMetadata(obj) {
            fmt.Println("Object %v invalid, skipping...", obj)
            continue
        }

        // write out the file using the template
        writeTemplate(obj)
    }}
}

// setCaps takes a given block of text and capitalizes the first letter.
func setCaps(text string) string {
    r, n := utf8.DecodeRuneInString(text)
    return string(unicode.ToUpper(r)) + text[n:]
}

// validateMsgMetadata does some simple checking to make sure that enough 
// data was parsed to form a valid MsgGenData object.
func validateMsgMetadata(obj *MsgGenData) bool {
    return !(obj.Exports == nil || len(obj.Exports) < 1 || 
        obj.NetMsg  == ""  || obj.MsgSig == "" || 
        obj.Package == "")
}

// validateStructComment applies a regex to test if a given comment line
// contains a NetMsg flag, and also captures and returns the signature data,
// if present.
func validateStructComment(text string) (string, bool) {
    matches := netMsgRegexp.FindStringSubmatch(text)

    if len(matches) < 2 {
        return "", false
    }

    return matches[1], true
}

// writeTemplate accepts a valid MsgGenData object, applies it to the template,
// and writes the results out to a new source file.
func writeTemplate(data *MsgGenData) {
    path := filepath.Join(data.Path, "msg" + data.NetMsg + ".go")

    f, err := os.Create(path)
    if err != nil {
        panic(err)
    }
    defer f.Close()

    tmpl, err := template.New(data.NetMsg).Parse(templateTxt)
    if err != nil { panic(err) }
    err = tmpl.Execute(f, data)
    if err != nil { panic(err) }

    fmt.Printf("\twrote: %s\n\n", path)
}


var templateTxt =
`//  ---------------------------------------------------------------------------
//
//  msg{{.NetMsg}}.go
//
//  This file is auto-generated by the net message code generator and should 
//  NOT be edited by hand unless you know what you are doing. Changes to the
//  source object definition will be automatically reflected in the this 
//  generated code the next time genproc is run.
//
//  -----------
package {{.Package}}

// External imports.
import (
    "github.com/xaevman/goat/core/net"
    "github.com/xaevman/goat/lib/buffer"
)

// Stdlin imports.
import (
    "errors"
    "fmt"
)

// Generated imports.
import ({{range $import := .Imports}}
    {{$import}}{{end}}
)

// {{.NetMsg}}Handler is an empty function container.
type {{.NetMsg}}Handler struct {}

// Close is called when a message signature is unregistered from a protocol.
func (this *{{.NetMsg}}Handler) Close() {}

// Init is called when the message signature is first registered in a protocol.
func (this *{{.NetMsg}}Handler) Init(proto *net.Protocol) {}

// DeserializeMsg is called by the protocol after an incoming network message has been 
// validated, decrypted, and uncompressed.
func (this *{{.NetMsg}}Handler) DeserializeMsg(msg *net.Msg, access byte) (interface{}, error) {
    var err error

    cursor := 0
    data   := msg.GetPayload()
    nMsg   := new({{.NetMsg}})
    {{range $export := .Exports}}
    nMsg.{{$export.Name}}, err = buffer.Read{{$export.Type}}(data, &cursor)
    if err != nil { return nil, err }
    {{end}}
    return nMsg, nil
}

// SerializeMsg is called by the protocol after a {{.NetMsg}} object has been validated,
// compressed, and encrypted, in order to prepare a network message for transmission.
func (this *{{.NetMsg}}Handler) SerializeMsg(data interface{}) (*net.Msg, error) {
    cursor      := 0
    nMsg, ok := data.(*{{.NetMsg}})
    if !ok {
        return nil, errors.New(fmt.Sprintf("Cannot serialize type %T", data))
    }

    dataLen := 0
    {{range $export := .Exports}}
    dataLen += buffer.Len{{$export.Type}}({{if eq $export.Type "String"}}nMsg.{{$export.Name}}{{end}}){{end}}

    dataBuffer := make([]byte, dataLen)
    {{range $export := .Exports}}
    buffer.Write{{$export.Type}}(nMsg.{{$export.Name}}, dataBuffer, &cursor){{end}}

    msg := net.NewMsg()
    msg.SetMsgType(this.Signature())
    msg.SetPayload(dataBuffer)

    return msg, nil
}

// Signature returns {{.NetMsg}}'s network signature ({{.MsgSig}}).
func (this *{{.NetMsg}}Handler) Signature() uint16 {
    return {{.MsgSig}}
}
`
