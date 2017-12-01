package tfdoc

import (
	"bufio"
	"bytes"
	"container/list"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync/atomic"
)

type STATUS int

const (
	_enum     STATUS = 1
	_tfObject        = 2
	_service         = 3
)

var tfObjectList = newTfObjects()
var tFObjectMap = make(map[string]*TFObject, 0)
var servcieList = make([]*Servcie, 0)
var typedefList = make([]string, 0)
var constList = make([]string, 0)
var enumList = make([]string, 0)
var i int32 = 0

type Servcie struct {
	ServcieBody string
}

//typedef
func isTypedef(line string) (isTypedef bool) {
	return strings.HasPrefix(line, "typedef ")
}

//const
func isConst(line string) (isConst bool) {
	return strings.HasPrefix(line, "const ")
}

//enum
func isEnum(line string) bool {
	return strings.HasPrefix(line, "enum ")
}

//struct
func getStruct(line string) (name string) {
	return _get(line, "struct ")
}

//service
func getService(line string) (name string) {
	return _get(line, "service ")
}

func _get(line, objname string) (name string) {
	line = strings.TrimSpace(line)
	line = strings.Replace(line, objname, "", -1)
	name = strings.Replace(line, "{", "", -1)
	name = strings.TrimSpace(name)
	return
}

//
func getField(line string, ls *list.List) {
	line = strings.TrimSpace(line)
	pattern := "^[1-9]{1,}\\s{0,}"
	b, err := regexp.Match(pattern, []byte(line))
	if b && err == nil {
		line = replaceSq(line, " ", ":", "map<", "list<", "set<", ">", "\t", ",", "optional", "required")
		hasEq := false
		if strings.Contains(line, "=") {
			hasEq = true
			line = replaceSq(line, "=")
		}
		ss := strings.Split(line, " ")
		length := len(ss) - 1
		if hasEq {
			length = length - 1
		}
		for i := 1; i < length; i++ {
			getType(ss[i], ls)
		}
	}
}

func replaceSq(line string, sq string, ss ...string) string {
	for _, s := range ss {
		line = strings.Replace(line, s, sq, -1)
	}
	return line
}

func getType(fieldType string, ls *list.List) {
	if fieldType == "" {
		return
	}
	if strings.Contains(fieldType, ".") {
		fieldType = strings.Split(fieldType, ".")[1]
	}
	if !isEqs(fieldType, "bool", "i8", "i16", "i32", "i64", "double", "binary", "string") {
		ls.PushFront(fieldType)
	}
}

//是否集合
func isCollections(fieldType string) bool {
	if strings.HasPrefix(fieldType, "map<") || strings.HasPrefix(fieldType, "list<") || strings.HasPrefix(fieldType, "set<") {
		return true
	}
	return false
}

func getStatus(line string) STATUS {
	if strings.HasPrefix(line, "struct") {
		return _tfObject
	}
	if strings.HasPrefix(line, "enum") {
		return _enum
	}
	if strings.HasPrefix(line, "exception") {
		return _tfObject
	}
	return 0
}

func isEqs(s string, fs ...string) bool {
	for _, f := range fs {
		if s == f {
			return true
		}
	}
	return false
}

func replaceTrim(line string, ss ...string) string {
	for _, s := range ss {
		line = strings.Replace(line, s, "", -1)
	}
	return line
}

func WalkDir(dirPth, suffix string) (err error) {
	suffix = strings.ToUpper(suffix)                                                     //忽略后缀匹配的大小写
	err = filepath.Walk(dirPth, func(filename string, fi os.FileInfo, err error) error { //遍历目录
		if err != nil { //忽略错误
			return err
		}
		if fi.IsDir() { // 忽略目录
			return nil
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			Readfile(filename)
		}
		return nil
	})
	return err
}

func Readfile(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	buf := bufio.NewReader(f)
	var tfobj *TFObject
	var servcieObj *Servcie
	var buffer bytes.Buffer
	var status STATUS = 0
	var isComment = false
	for {
		line, err := buf.ReadString('\n')
		if err != nil && strings.TrimSpace(line) == "" {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" || line == "\n" {
			continue
		}
		if status == 0 && (strings.Contains(line, "namespace ") || strings.Contains(line, "include ")) {
			continue
		}
		if strings.HasPrefix(line, "/**") && !strings.HasSuffix(line, "*/") {
			isComment = true
		}
		buffer.WriteString(line + "\n")
		if isComment && !strings.HasSuffix(line, "*/") {
			continue
		}
		if isComment && strings.HasSuffix(line, "*/") {
			isComment = false
			continue
		}

		b := isTypedef(line)
		if b {
			typedefList = append(typedefList, buffer.String())
			buffer.Reset()
			continue
		}
		b = isConst(line)
		if b {
			constList = append(constList, buffer.String())
			buffer.Reset()
			continue
		}
		if strings.HasPrefix(line, "enum ") {
			status = _enum
			continue
		}
		if getStatus(line) > 0 {
			status = getStatus(line)
			tfobj = new(TFObject)
			tfobj.name = getStruct(line)
			tfobj.dependency = list.New()
			atomic.AddInt32(&i, 1)
			tfobj.score = i
			continue
		}
		if strings.HasPrefix(line, "service ") {
			status = _service
			servcieObj = new(Servcie)
			continue
		}
		if strings.HasPrefix(line, "}") {
			switch status {
			case _enum:
				enumList = append(enumList, buffer.String())
			case _tfObject:
				body := fmt.Sprint("/**", f.Name(), "*/\n", buffer.String())
				tfobj.body = body
				tfObjectList.add(tfobj)
				tFObjectMap[tfobj.name] = tfobj
			case _service:
				body := fmt.Sprint("/**", f.Name(), "*/\n", buffer.String())
				servcieObj.ServcieBody = body
				servcieList = append(servcieList, servcieObj)
			default:
			}
			buffer.Reset()
			status = 0
			continue
		}
		if status == _tfObject {
			getField(line, tfobj.dependency)
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
	fmt.Println("file:", f.Name())
	return nil
}

func score(tfobj *TFObject) {
	if tfobj.dependency.Len() > 0 {
		for e := tfobj.dependency.Front(); e != nil; e = e.Next() {
			name := e.Value.(string)
			if name != "" {
				if v, ok := tFObjectMap[name]; ok {
					v.score = v.score + tfobj.score
					score(v)
				}
			}
		}
	}
}

func parse() {
	//	fmt.Println("object:", tfObjectList.Len())
	fmt.Println("distinct object:", len(tFObjectMap))
	fmt.Println("servcie:", len(servcieList))
	fmt.Println("typedef:", len(typedefList))
	fmt.Println("const:", len(constList))
	fmt.Println("enum:", len(enumList))
	for _, tfobj := range *tfObjectList {
		score(tfobj)
	}
}

func CreateNewThrifFile(filename string, namespace string) {
	parse()
	var buffer bytes.Buffer
	if namespace != "" {
		buffer.WriteString(namespace)
		buffer.WriteString("\n")
	}
	for _, s := range typedefList {
		buffer.WriteString(s)
	}
	buffer.WriteString("\n")
	for _, s := range constList {
		buffer.WriteString(s)
	}
	buffer.WriteString("\n")

	for _, s := range enumList {
		buffer.WriteString(s)
	}
	buffer.WriteString("\n")
	sort.Sort(*tfObjectList)

	for _, v := range *tfObjectList {
		buffer.WriteString(v.body)
	}
	buffer.WriteString("\n")

	for _, s := range servcieList {
		buffer.WriteString(s.ServcieBody)
	}

	os.Remove(filename)
	err := ioutil.WriteFile(filename, buffer.Bytes(), 777)
	if err != nil {
		fmt.Println(err.Error())
	}
}
