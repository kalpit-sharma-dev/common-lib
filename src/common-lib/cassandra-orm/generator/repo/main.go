package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

const (
	templatePath       = "/src/gitlab.kksharmadevdev.com/platform/platform-common-lib/src/cassandra-orm/template/entity_repository.go"
	selectByViewTables = `
// Get{{.ViewTable}} returns item from '{{.TableName}}' table
func (r *{{.Entity}}Repository) Get{{.ViewTable}}(keyCols ...interface{}) (*{{.Entity}}, error) {
	item := new({{.Entity}})
	if err := r.base.GetFromTable({{.ConstName}}, item, keyCols...); err != nil {
		return nil, err
	}
	return item, nil
}

// All{{.ViewTable}} returns all items from '{{.TableName}}' table
func (r *{{.Entity}}Repository) All{{.ViewTable}}(keyCols ...interface{}) ([]*{{.Entity}}, error) {
	var items []*{{.Entity}}
	if err := r.base.AllFromTable({{.ConstName}}, &items, keyCols...); err != nil {
		return nil, err
	}
	return items, nil
}
`
)

type (
	viewsTableTempl struct {
		ViewTable string
		TableName string
		Entity    string
		ConstName string
		Keys      []string
	}

	viewTable struct {
		Name string
		Keys []string
	}
)

// Usage (in one line):
// go:generate repo
// 	-package=model
// 	-out=./cat_repository_gen.go
// 	-table=cats
// 	-keys="ID,Age"
//  -viewTables=[cats_by_age:"Age,ID"]
// "entity=cat entities=cats"

func main() {
	var (
		pkgName    = flag.String("package", "", "package name for generated files")
		out        = flag.String("out", "", "file to save output to instead of stdout")
		idType     = flag.String("type", "gocql.UUID", "primary key type for generated repo")
		viewTables = flag.String("viewTables", "", "view table names (instead of views) for generated repo")
		table      = flag.String("table", "", "table name for generated repo")
		keys       = flag.String("keys", "", "list of keys for generated repo")
	)
	flag.Parse()

	in := os.Getenv("GOPATH") + templatePath
	keyFields := getKeys(*keys)
	allSubst, entityName := populateSubstitutions(flag.Args(), *idType)

	var baseViewTables []*viewTable
	if *viewTables != "" {
		baseViewTables = getHelpTables(*viewTables)
		for _, t := range baseViewTables {
			fmt.Printf("%#v\n", t)
		}
	}

	fmt.Println("template:", in)
	fmt.Println("package:", *pkgName)
	fmt.Println("out:", *out)
	fmt.Println("table:", *table)
	fmt.Println("type:", *idType)
	fmt.Println("entity:", entityName)
	fmt.Println("key(s):", strings.Join(keyFields, ","))
	fmt.Println("substitution(s):", allSubst)
	fmt.Println()

	args := []string{"gen", allSubst}
	execute(&in, out, pkgName, args)

	content := postProcess(*out, entityName, *table, getKeyConstants(keyFields), baseViewTables)
	content = makeGoFmt(content)
	writeToFile(*out, content)
}

func getHelpTables(tablesInfo string) []*viewTable {
	tablesInfo = strings.TrimLeft(strings.TrimRight(tablesInfo, "]"), "[")
	tables := strings.Split(tablesInfo, ";")
	result := make([]*viewTable, len(tables))

	for i, table := range tables {
		name, keyFields := getTableNameAndKeys(table)
		result[i] = &viewTable{
			Name: name,
			Keys: keyFields,
		}
	}

	return result
}

func getKeys(keys string) []string {
	keys = strings.Trim(keys, "\"")
	return strings.Split(keys, ",")
}

func getKeyConstants(keys []string) []string {
	for i, key := range keys {
		keys[i] = key + "Column"
	}
	return keys
}

func getTableNameAndKeys(tableInfo string) (string, []string) {
	parts := strings.Split(tableInfo, ":")
	if len(parts) != 2 {
		fmt.Printf("Error: table defined wrongly: %q", tableInfo)
	}
	return parts[0], getKeyConstants(getKeys(parts[1]))
}

func populateSubstitutions(args []string, idType string) (allSubstitutions, entityName string) {
	substSlice := make([]string, 0)
	for _, arg := range args {
		parts := strings.Split(arg, " ")

		for _, part := range parts {
			subParts := strings.Split(part, "=")
			substSlice = append(substSlice, part, strings.Title(subParts[0])+"="+strings.Title(subParts[1]))
			if subParts[0] == "entity" {
				entityName = subParts[1]
			}

		}
	}
	allSubstitutions = strings.Join(substSlice, " ") + " IDType=" + idType
	return
}

func postProcess(outFileName, entityName, table string, keys []string, viewTables []*viewTable) []byte {
	outFile, err := os.Open(filepath.Clean(outFileName))
	if err != nil {
		panic(err)
	}
	defer func() {
		checkAndLog(outFile.Close())
	}()

	tableNamePrefix := fmt.Sprintf(`const %sBaseTableName = "`, entityName)
	keyColumnsPrefix := fmt.Sprintf(`%sKeyColumns = []string{"`, entityName)
	viewTablesPrefix := fmt.Sprintf(`%sViewTables = map[string][]string{`, entityName)

	tableTempls := make([]*viewsTableTempl, len(viewTables))
	for i, viewTable := range viewTables {
		tableTempls[i] = getViewsTableTempl(viewTable, entityName)
	}

	buffer := &bytes.Buffer{}
	fileScanner := bufio.NewScanner(outFile)
	for fileScanner.Scan() {
		text := fileScanner.Text()
		trimmedLeft := strings.TrimLeft(text, " \t")

		switch {
		case strings.HasPrefix(text, tableNamePrefix):
			text = getConstants(entityName, table, tableTempls)
		case strings.HasPrefix(trimmedLeft, keyColumnsPrefix):
			text = fmt.Sprintf("\t"+`%sKeyColumns = []string{%s}`, entityName, strings.Join(keys, `, `))
		case strings.HasPrefix(trimmedLeft, viewTablesPrefix):
			tables := make([]string, len(tableTempls))
			for i, tableTempl := range tableTempls {
				tables[i] = tableTempl.String()
			}
			tablesValue := strings.Join(tables, ",\n")
			if len(tables) != 0 {
				tablesValue = fmt.Sprintf("\n%s,\n", tablesValue)
			}
			text = fmt.Sprintf("\t"+`%sViewTables = map[string][]string{%s}`, entityName, tablesValue)
		}
		_, err = buffer.WriteString(text + "\n")
		if err != nil {
			panic(err)
		}
	}
	err = addViewFunctions(buffer, tableTempls)
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
}

func getConstants(entityName, table string, tableTempls []*viewsTableTempl) string {
	if len(tableTempls) == 0 {
		return fmt.Sprintf(`const %sBaseTableName = "%s"`, entityName, table)
	}
	return fmt.Sprintf(`const (%sBaseTableName = "%s"%s)`, entityName, table, getConstantsList(tableTempls))
}

func addViewFunctions(w io.Writer, tableTempls []*viewsTableTempl) error {
	if len(tableTempls) == 0 {
		return nil
	}
	tmpl, err := template.New("template").Parse(selectByViewTables)
	if err != nil {
		return err
	}

	for _, tableTempl := range tableTempls {
		err = tmpl.Execute(w, tableTempl)
		if err != nil {
			return err
		}
	}
	return nil
}

func getViewsTableTempl(viewTable *viewTable, entityName string) *viewsTableTempl {
	viewTableName := getViewTableName(viewTable.Name)
	return &viewsTableTempl{
		ViewTable: viewTableName,
		TableName: viewTable.Name,
		Entity:    strings.Title(entityName),
		ConstName: lowercaseFirst(viewTableName),
		Keys:      viewTable.Keys,
	}
}

func getViewTableName(tableName string) string {
	tempName := strings.Replace(tableName, "_", " ", -1)
	tempName = strings.Title(tempName)
	return strings.Replace(tempName, " ", "", -1)
}

func makeGoFmt(content []byte) []byte {
	result, err := format.Source(content)
	if err != nil {
		panic(err)
	}
	return result
}

func writeToFile(outFileName string, content []byte) {
	path, _ := filepath.Split(outFileName)
	tempFile, err := ioutil.TempFile(path, "tmp_gen_")
	if err != nil {
		panic(err)
	}
	defer func() {
		checkAndLog(tempFile.Close())
		checkAndLog(os.Rename(tempFile.Name(), outFileName))
	}()

	w := bufio.NewWriter(tempFile)
	defer func() {
		checkAndLog(w.Flush())
	}()
	_, err = w.Write(content)
	if err != nil {
		panic(err)
	}
}

func checkAndExit(code int, err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(code)
	}
}

func checkAndLog(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func (t *viewsTableTempl) String() string {
	return fmt.Sprintf(`%s: {%s}`, t.ConstName, strings.Join(t.Keys, `, `))
}

func (t *viewsTableTempl) StringConst() string {
	return fmt.Sprintf(`%s = "%s"`, t.ConstName, t.TableName)
}

func getConstantsList(tableTempls []*viewsTableTempl) string {
	constants := make([]string, len(tableTempls))
	for i, tableTempl := range tableTempls {
		constants[i] = tableTempl.StringConst()
	}
	return "\n" + strings.Join(constants, "\n")
}

func lowercaseFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}
