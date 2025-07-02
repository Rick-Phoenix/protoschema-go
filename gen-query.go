package protoschema

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	u "github.com/Rick-Phoenix/goutils"
	"github.com/Rick-Phoenix/protoschema/db"
	"github.com/labstack/gommon/log"
	_ "modernc.org/sqlite"
)

type Subquery struct {
	Method    string
	ParamName string
}

type SubqueryData struct {
	Method     string
	ParamName  string
	VarName    string
	ReturnType string
}

type Value struct {
	PkgPath string
}

type QuerySchema struct {
	Name       string
	Subqueries []Subquery
	OutType    any
}

type QueryData struct {
	Name            string
	FunctionParams  map[string]string
	OutType         string
	OutTypeFields   []string
	Subqueries      []SubqueryData
	MakeParamStruct bool
	FuncParamName   string
	FuncParamType   string
	Imports         *u.Set[string]
}

func (p *ProtoPackage) makeQuery() {
	tmpl := p.tmpl
	database, err := sql.Open("sqlite", "db/database.sqlite3?_time_format=sqlite")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	qmap := make(map[string]string)
	qmap["GetUser"] = "userId"

	store := db.NewStore(database)
	val := reflect.TypeOf(store.Queries)

	imports := u.NewSet[string]()
	queryData := QueryData{Name: "GetUserWithPosts", FunctionParams: make(map[string]string), Imports: imports}

	subqueries := []Subquery{{"GetUser", "userId"}, {"GetPostsFromUserId", ""}}

	subQueriesData := []SubqueryData{}

	for _, subQ := range subqueries {
		subQData := SubqueryData{Method: subQ.Method}
		method, _ := val.MethodByName(subQ.Method)

		fmt.Printf("Method Name: %+v\n", method.Name)
		if method.Type.NumIn() >= 3 {
			secondParam := method.Type.In(2)
			if secondParam.Kind() == reflect.Struct {
				subQData.ParamName = secondParam.Name()
				queryData.FunctionParams[secondParam.Name()] = secondParam.String()
				for i := range secondParam.NumField() {
					field := secondParam.Field(i)
					fmt.Printf("Param %d in %s: %+v %+v\n", i, method.Name, field.Name, field.Type.Name())
				}
			} else {
				fmt.Printf("Single Param: %+v\n", secondParam.Name())
				if v, ok := qmap[method.Name]; ok {
					queryData.FunctionParams[v] = secondParam.Name()
					subQData.ParamName = v
				}
			}
		}

		if method.Type.NumOut() > 0 {
			out := method.Type.Out(0)
			outElem := out.Elem()
			outShortType := outElem.Name()
			outLongType := out.String()
			outPkgPath := outElem.PkgPath()
			if out.Kind() == reflect.Slice {
				outShortType = outElem.Elem().Name() + "s"
				outPkgPath = outElem.Elem().PkgPath()
			}
			outShortLower := u.Uncapitalize(outShortType)
			subQData.VarName = outShortLower
			subQData.ReturnType = outLongType
			fmt.Printf("Outshort: %+v\n", outShortType)
			fmt.Printf("Outshortlower: %+v\n", outShortLower)
			fmt.Printf("Outlong: %+v\n", outLongType)
			fmt.Printf("Outpkgpath: %+v\n", outPkgPath)
		}

		subQueriesData = append(subQueriesData, subQData)
	}

	queryData.Subqueries = subQueriesData

	outType := &db.UserWithPosts{}

	outModel := reflect.TypeOf(outType).Elem()

	queryData.OutType = outModel.String()

	for i := range outModel.NumField() {
		field := outModel.Field(i)

		fmt.Printf("Outfield %d: %+v\n", i, field.Name)
		queryData.OutTypeFields = append(queryData.OutTypeFields, field.Name)
	}

	if len(queryData.FunctionParams) > 1 {
		queryData.MakeParamStruct = true
		queryData.FuncParamName = "params"
		queryData.FuncParamType = queryData.Name + "Params"
	} else {
		for name, typ := range queryData.FunctionParams {
			queryData.FuncParamName = name
			queryData.FuncParamType = typ
		}
	}

	fmt.Printf("DEBUG: %+v\n", queryData)

	outputPath := "gen/tttestquery.go"

	var outputBuffer bytes.Buffer
	if err := tmpl.ExecuteTemplate(&outputBuffer, "multiQuery", queryData); err != nil {
		fmt.Printf("Failed to execute template: %s", err.Error())
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		fmt.Println("Error dir")
	}
	if err := os.WriteFile(outputPath, outputBuffer.Bytes(), 0644); err != nil {
		fmt.Println("Error writing the file")
	}
}
