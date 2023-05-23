package utils

//Package utils is to convert interface type to specific type

import (
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"crypto/md5"
	"encoding/hex"

	"github.com/google/uuid"
	errorCodes "gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/errorCodePair"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/constants"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/plugin/protocol"
)

// ToString converts an interface type to string
func ToString(v interface{}) string {
	t, ok := v.(string)
	if ok {
		return t
	}
	return ""
}

// ToTime converts an interface type to time
func ToTime(v interface{}) time.Time {
	t, ok := v.(time.Time)
	if ok {
		return t
	}
	return time.Time{}
}

// ToPtrTime converts an interface type to *time
func ToPtrTime(v interface{}) *time.Time {
	t, ok := v.(time.Time)
	if ok {
		if t.IsZero() {
			return nil
		}
		return &t
	}
	return nil
}

// ToIntArray converts an interface type to int arr
func ToIntArray(v interface{}) []int {
	t, ok := v.([]int)
	if ok {
		return t
	}
	return []int{}
}

// ToInt64 converts an interface type to int64
// interface{} holding an int will not be type casted to int64 and will return 0 as the result
func ToInt64(v interface{}) int64 {
	t, ok := v.(int64)
	if ok {
		return t
	}
	return 0
}

// ToUint64 converts an interface type to uint64
// interface{} holding an int will not be type casted to uint64 and will return 0 as the result
func ToUint64(v interface{}) uint64 {
	return uint64(ToInt64(v))
}

// ToInt converts an interface type to int
func ToInt(v interface{}) int {
	t, ok := v.(int)
	if ok {
		return t
	}
	return 0
}

// ToFloat64 converts an interface type to float64
func ToFloat64(v interface{}) float64 {
	t, ok := v.(float64)
	if ok {
		return t
	}
	return 0
}

// ToBool converts an interface type to bool
func ToBool(v interface{}) bool {
	t, ok := v.(bool)
	if ok {
		return t
	}
	return false
}

// ToStringArray converts an interface type to string array
func ToStringArray(v interface{}) []string {
	t, ok := v.([]string)
	if ok {
		return t
	}
	return []string{}
}

// ToByteArray converts an interface type to byte array
func ToByteArray(v interface{}) []byte {
	t, ok := v.([]byte)
	if ok {
		return t
	}
	return []byte{}
}

// ToStringMap converts an interface type to map[string]string
func ToStringMap(v interface{}) map[string]string {
	t, ok := v.(map[string]string)
	if ok {
		return t
	}
	return nil
}

// GetTransactionID generates new transactionid
func GetTransactionID() string {
	return uuid.New().String()
}

// GetTransactionIDFromResponse retrieves transactionid from the Response header
func GetTransactionIDFromResponse(res *http.Response) string {
	if res == nil {
		return ""
	}
	return res.Header.Get(string(protocol.HdrTransactionID))
}

// GetTransactionIDFromRequest retrieves transactionID from the Request header
func GetTransactionIDFromRequest(req *http.Request) string {
	if req == nil {
		return GetTransactionID()
	}
	value := GetValueFromRequestHeader(req, protocol.HdrTransactionID)
	if value != "" {
		return value

	}
	value = req.Header.Get(constants.TransactionID)
	if value != "" {
		return value
	}
	return GetTransactionID()
}

// GetQueryValuesFromRequest to get query values from request for given filter
func GetQueryValuesFromRequest(req *http.Request, filter string) []string {
	queryValues := req.URL.Query()
	if _, ok := queryValues[filter]; ok {
		return queryValues[filter]
	}
	return []string{}
}

// GetChecksumFromRequest retrives MD5 from the request header
func GetChecksumFromRequest(req *http.Request) string {
	return GetValueFromRequestHeader(req, protocol.HdrContentMD5)
}

// GetValueFromRequestHeader retrieves header value for given Key from the Request header
func GetValueFromRequestHeader(req *http.Request, key protocol.HeaderKey) string {
	if req == nil {
		return ""
	}

	return req.Header.Get(string(key))
}

// GetChecksum is a function to calculate MD5 hash value
func GetChecksum(message []byte) string {
	hasher := md5.New()
	hasher.Write(message) //nolint
	return hex.EncodeToString(hasher.Sum(nil))
}

// ValidateMessage checks if message is corrupted or not
func ValidateMessage(message []byte, receievedChecksum string) (bool, string) {
	hashValue := GetChecksum(message)
	if receievedChecksum != "" && hashValue != receievedChecksum {
		return false, hashValue
	}
	return true, hashValue
}

// changes for agent autoupdate error standardization START HERE. To be refactored as per common-lib standards for comming rollouts

// DetermineErrorCodePair for autoupdate failures
func DetermineErrorCodePair(errMsg string) (mainError, subError string) {
	//determine errors code pairs
	if strings.Contains(errMsg, errorCodes.FileSystem) {
		mainError, subError = errorCodes.FileSystem, determineFileSystemErrorCodes(errMsg)
	} else if strings.Contains(errMsg, errorCodes.Network) {
		mainError, subError = errorCodes.Network, determineNetworkErrorCodes(errMsg)
	} else if strings.Contains(errMsg, errorCodes.Download) {
		mainError, subError = errorCodes.Download, determineDownloadErrorCodes(errMsg)
	} else if strings.Contains(errMsg, errorCodes.Internal) {
		mainError, subError = errorCodes.Internal, determineInternalErrorCodes(errMsg)
	}
	if subError == "" {
		mainError, subError = errorCodes.Internal, errorCodes.Operational
	}
	return
}

func determineFileSystemErrorCodes(errMsg string) string {
	if strings.Contains(errMsg, errorCodes.Database) {
		return errorCodes.Database
	} else if strings.Contains(errMsg, errorCodes.AccessDenied) {
		return errorCodes.AccessDenied
	} else if strings.Contains(errMsg, errorCodes.Diskfull) {
		return errorCodes.Diskfull
	} else if strings.Contains(errMsg, errorCodes.FileNotFound) {
		return errorCodes.FileNotFound
	}
	return ""
}

func determineNetworkErrorCodes(errMsg string) string {
	if strings.Contains(errMsg, errorCodes.Connection) {
		return errorCodes.Connection
	} else if strings.Contains(errMsg, errorCodes.Proxy) {
		return errorCodes.Proxy
	}
	return ""
}

func determineDownloadErrorCodes(errMsg string) string {
	if strings.Contains(errMsg, errorCodes.ChecksumValidationFailed) {
		return errorCodes.ChecksumValidationFailed
	}
	return ""
}

func determineInternalErrorCodes(errMsg string) string {
	if strings.Contains(errMsg, errorCodes.FalseAlarm) {
		return errorCodes.FalseAlarm
	} else if strings.Contains(errMsg, errorCodes.ProcessRunning) {
		return errorCodes.ProcessRunning
	} else if strings.Contains(errMsg, errorCodes.InstallFailure) {
		return errorCodes.InstallFailure
	}
	return ""
}

// changes for agent autoupdate error standardization END HERE.

// EqualFold is to check if two strings array are same with case insensitive
func EqualFold(source, target []string) bool {
	sort.Slice(source, func(i, j int) bool { return strings.ToLower(source[i]) < strings.ToLower(source[j]) })
	sort.Slice(target, func(i, j int) bool { return strings.ToLower(target[i]) < strings.ToLower(target[j]) })

	lSource := len(source)
	lTarget := len(target)

	if lSource != lTarget {
		return false
	}

	for i := 0; i < lSource; i++ {
		if !strings.EqualFold(source[i], target[i]) {
			return false
		}
	}

	return true
}

// Int64SliceToStringSlice converts int64 slice to string slice
func Int64SliceToStringSlice(in []int64) (out []string) {
	for _, v := range in {
		s := strconv.Itoa(int(v))
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// Change is a struct that can hold info about one change between two entities
type Change struct {
	Path []string
	From interface{}
	To   interface{}
}

// Difference returns differences between any two entities
func Difference(x, y interface{}) []Change {
	changesMap := make(map[string]Change)

	v1 := reflect.ValueOf(x)
	v2 := reflect.ValueOf(y)

	diff(v1, v2, []string{}, &changesMap, false)
	tmpChangesMap := make(map[string]Change)
	diff(v2, v1, []string{}, &tmpChangesMap, false)

	changes := []Change{}
	for _, change := range changesMap {
		changes = append(changes, change)
	}
	tmpChanges := []Change{}
	for _, change := range tmpChangesMap {
		tmpChanges = append(tmpChanges, change)
	}

	mergeChanges(&changes, tmpChanges)
	return changes
}

func diff(v1, v2 reflect.Value, path []string, changes *map[string]Change, lastFieldHasTag bool) {

	if !v1.IsValid() || !v2.IsValid() {
		return
	}

	if v1.Type() != v2.Type() {
		diffWithNil(v1, path, changes, lastFieldHasTag)
		return
	}

	if !v1.IsValid() || !v2.IsValid() {
		newChange := Change{
			Path: path,
			From: v1,
			To:   v2,
		}
		(*changes)[pathSliceToString(path)] = newChange
		return
	}

	switch v1.Kind() {
	case reflect.Array:
		for i := 0; i < v1.Len(); i++ {
			newPath := append(path, fmt.Sprint(i))
			if i >= v2.Len() {
				newChange := Change{
					Path: newPath,
					From: v1.Index(i).Interface(),
					To:   nil,
				}
				(*changes)[pathSliceToString(newPath)] = newChange
				continue
			}
			diff(v1.Index(i), v2.Index(i), newPath, changes, lastFieldHasTag)
		}
	case reflect.Slice:
		if v1.Pointer() == v2.Pointer() {
			return
		}
		for i := 0; i < v1.Len(); i++ {
			newPath := append(path, fmt.Sprint(i))
			if i >= v2.Len() {
				newChange := Change{
					Path: newPath,
					From: v1.Index(i).Interface(),
					To:   nil,
				}
				(*changes)[pathSliceToString(newPath)] = newChange
				continue
			}
			diff(v1.Index(i), v2.Index(i), newPath, changes, lastFieldHasTag)
		}
		return
	case reflect.Interface:
		if v1.IsNil() || v2.IsNil() {
			if v1.IsNil() != v2.IsNil() {
				newChange := Change{
					Path: path,
					From: v1,
					To:   v2,
				}
				(*changes)[pathSliceToString(path)] = newChange
				return
			}
		}
		diff(v1.Elem(), v2.Elem(), path, changes, lastFieldHasTag)
	case reflect.Ptr:
		if v1.Pointer() == v2.Pointer() {
			return
		}
		newPath := append(path, v1.Type().Name())
		diff(v1.Elem(), v2.Elem(), newPath, changes, lastFieldHasTag)
	case reflect.Struct:
		switch v := v1.Interface().(type) {
		case time.Time:
			v.String()
			if !reflect.DeepEqual(v1.Interface(), v2.Interface()) {
				newChange := Change{
					Path: path,
					From: v1.Interface(),
					To:   v2.Interface(),
				}
				(*changes)[pathSliceToString(path)] = newChange
			}
		default:
			for i, n := 0, v1.NumField(); i < n; i++ {
				fieldName := v1.Type().Field(i).Name
				tagName := strings.Split(v1.Type().Field(i).Tag.Get("json"), ",")[0]
				hasTag := false
				if tagName != "" {
					hasTag = true
					fieldName = tagName
				}
				newPath := append(path, fieldName)
				fieldExist := false
				if _, ok := (*changes)[pathSliceToString(newPath)]; ok {
					fieldExist = true
				}

				if i >= v2.NumField() {
					newChange := Change{
						Path: path,
						From: v1.Field(i),
						To:   nil,
					}
					if fieldExist {
						if hasTag {
							// if has tag override existing value
							(*changes)[pathSliceToString(path)] = newChange
						}
						// otherwise if it field exists and this field doesn't has a tag then ignore
					} else {
						// if field path doesn't exist add it
						(*changes)[pathSliceToString(path)] = newChange
					}
					continue
				}
				diff(v1.Field(i), v2.Field(i), newPath, changes, hasTag)
			}
		}
		return
	case reflect.Map:
		if v1.IsNil() != v2.IsNil() {
			newChange := Change{
				Path: path,
				From: v1,
				To:   v2,
			}
			(*changes)[pathSliceToString(path)] = newChange
			return
		}
		if v1.Pointer() == v2.Pointer() {
			return
		}
		for _, k := range v1.MapKeys() {
			val1 := v1.MapIndex(k)
			val2 := v2.MapIndex(k)
			newPath := append(path, k.String())
			if !val1.IsValid() {
				diffWithNil(val2, newPath, changes, lastFieldHasTag)
			} else if !val2.IsValid() {
				diffWithNil(val1, newPath, changes, lastFieldHasTag)
			} else {
				diff(val1, val2, newPath, changes, lastFieldHasTag)
			}
		}
		return
	case reflect.Func:
		if v1.IsNil() && v2.IsNil() {
			return
		}
		newChange := Change{
			Path: path,
			From: v1,
			To:   v2,
		}
		(*changes)[pathSliceToString(path)] = newChange
		return
	default:
		fieldExist := false
		if _, ok := (*changes)[pathSliceToString(path)]; ok {
			fieldExist = true
		}

		addChange := false
		if fieldExist {
			if lastFieldHasTag {
				// if has tag override existing value
				addChange = true
			}
			// otherwise if it field exists and this field doesn't has a tag then ignore
		} else {
			// if field path doesn't exist add it
			addChange = true
		}

		if v1.CanInterface() && v2.CanInterface() {
			if v1.Interface() != v2.Interface() {
				if addChange {
					newChange := Change{
						Path: path,
						From: v1.Interface(),
						To:   v2.Interface(),
					}
					(*changes)[pathSliceToString(path)] = newChange
				}
			}
		} else if v1.CanAddr() && v2.CanAddr() {
			if addChange {
				newChange := Change{
					Path: path,
					From: v1.Addr(),
					To:   v2.Addr(),
				}
				(*changes)[pathSliceToString(path)] = newChange
			}
		}
		return
	}
}

func diffWithNil(v reflect.Value, path []string, changes *map[string]Change, lastFieldHasTag bool) {

	switch v.Kind() {
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			newPath := append(path, fmt.Sprint(i))
			diffWithNil(v.Index(i), newPath, changes, lastFieldHasTag)
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			newPath := append(path, fmt.Sprint(i))
			diffWithNil(v.Index(i), newPath, changes, lastFieldHasTag)
		}
		return
	case reflect.Interface:
		diffWithNil(v.Elem(), path, changes, lastFieldHasTag)
	case reflect.Ptr:
		newPath := append(path, v.Type().Name())
		diffWithNil(v.Elem(), newPath, changes, lastFieldHasTag)
	case reflect.Struct:
		switch ty := v.Interface().(type) {
		case time.Time:
			ty.String()
			if v.Interface() != nil {
				newChange := Change{
					Path: path,
					From: v.Interface(),
					To:   nil,
				}
				(*changes)[pathSliceToString(path)] = newChange
			}
		default:
			for i, n := 0, v.NumField(); i < n; i++ {
				fieldName := v.Type().Field(i).Name
				tagName := strings.Split(v.Type().Field(i).Tag.Get("json"), ",")[0]
				hasTag := false
				if tagName != "" {
					hasTag = true
					fieldName = tagName
				}
				newPath := append(path, fieldName)
				diffWithNil(v.Field(i), newPath, changes, hasTag)
			}
		}
		return
	case reflect.Map:
		for _, k := range v.MapKeys() {
			val1 := v.MapIndex(k)
			newPath := append(path, k.String())
			diffWithNil(val1, newPath, changes, lastFieldHasTag)
		}
		return
	case reflect.Func:
		newChange := Change{
			Path: path,
			From: v,
			To:   nil,
		}
		(*changes)[pathSliceToString(path)] = newChange
		return
	default:
		fieldExist := false
		if _, ok := (*changes)[pathSliceToString(path)]; ok {
			fieldExist = true
		}

		addChange := false
		if fieldExist {
			if lastFieldHasTag {
				// if has tag override existing value
				addChange = true
			}
			// otherwise if it field exists and this field doesn't has a tag then ignore
		} else {
			// if field path doesn't exist add it
			addChange = true
		}

		if addChange {
			newChange := Change{
				Path: path,
				From: v.Interface(),
				To:   nil,
			}
			(*changes)[pathSliceToString(path)] = newChange
		}
		return
	}
}

func mergeChanges(changes *[]Change, newChanges []Change) {
	var mergedChanges []Change

	for _, c := range *changes {
		for _, nc := range newChanges {
			if reflect.DeepEqual(c.Path, nc.Path) {
				c.To = nc.From
				break
			}
		}

		if !reflect.DeepEqual(c.From, c.To) {
			mergedChanges = append(mergedChanges, c)
		}
	}

	for _, nc := range newChanges {
		foundInChanges := false
		for _, c := range *changes {
			if reflect.DeepEqual(c.Path, nc.Path) {
				foundInChanges = true
				break
			}
		}

		if !foundInChanges {
			tmpFrom := nc.From
			nc.From = nc.To
			nc.To = tmpFrom
			mergedChanges = append(mergedChanges, nc)
		}
	}

	*changes = mergedChanges
}

func pathSliceToString(path []string) string {
	var pathStr string = ""
	for _, p := range path {
		pathStr = pathStr + "/" + p
	}
	return pathStr
}

func GetStructFieldsTypes(entity interface{}) map[string]string {
	fieldsTypesMap := make(map[string]string)
	getStructFieldsTypesMap(reflect.ValueOf(entity), []string{}, &fieldsTypesMap)
	return fieldsTypesMap
}

func getStructFieldsTypesMap(entity reflect.Value, path []string, fieldsMap *map[string]string) {

	if !entity.IsValid() {
		return
	}

	switch entity.Kind() {
	case reflect.Array:
		(*fieldsMap)[pathSliceToString(path)] = entity.Type().String()
		for i := 0; i < entity.Len(); i++ {
			newPath := append(path, fmt.Sprint(i))
			getStructFieldsTypesMap(entity.Index(i), newPath, fieldsMap)
		}
	case reflect.Slice:
		(*fieldsMap)[pathSliceToString(path)] = entity.Type().String()
		for i := 0; i < entity.Len(); i++ {
			newPath := append(path, fmt.Sprint(i))
			getStructFieldsTypesMap(entity.Index(i), newPath, fieldsMap)
		}
		return
	case reflect.Interface:
		if entity.IsNil() {
			return
		}
		getStructFieldsTypesMap(entity.Elem(), path, fieldsMap)
	case reflect.Ptr:
		newPath := append(path, entity.Type().Name())
		getStructFieldsTypesMap(entity.Elem(), newPath, fieldsMap)
	case reflect.Struct:
		for i, n := 0, entity.NumField(); i < n; i++ {
			fieldName := entity.Type().Field(i).Name
			tagName := strings.Split(entity.Type().Field(i).Tag.Get("json"), ",")[0]
			if tagName != "" {
				fieldName = tagName
			}
			newPath := append(path, fieldName)
			(*fieldsMap)[pathSliceToString(path)] = entity.Type().String()
			getStructFieldsTypesMap(entity.Field(i), newPath, fieldsMap)
		}
		return
	case reflect.Map:
		if entity.IsNil() {
			return
		}
		for _, k := range entity.MapKeys() {
			val1 := entity.MapIndex(k)
			newPath := append(path, k.String())
			(*fieldsMap)[pathSliceToString(path)] = entity.Type().String()
			getStructFieldsTypesMap(val1, newPath, fieldsMap)
		}
		return
	case reflect.Func:
		if entity.IsNil() {
			return
		}
		(*fieldsMap)[pathSliceToString(path)] = entity.Type().String()
		return
	default:
		if entity.CanInterface() || entity.CanAddr() {
			(*fieldsMap)[pathSliceToString(path)] = entity.Type().String()
		}
		return
	}
}
