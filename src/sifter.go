package api

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

type sifterItem struct {
	index int    // 索引值 [0, reflect.Valueof(s).NumField())
	field string // 结构体域名称

	isAnonymous bool          // 结构体嵌套的是否是匿名域（anonymous struct field）
	embedded    *cachedSifter // 结构体嵌套的间接引用

	alias       string // 序列化时采取的别名
	isOmitEmpty bool   // json 序列化选项（omitempty）

	cLevel int // confidential level（保密级别）
}

type cachedSifter struct {
	sifterItems []*sifterItem
}

var sifterCache struct {
	sync.RWMutex
	m map[reflect.Type]cachedSifter
}

type sifterItemCtx struct {
	rv reflect.Value // si索引指向的所属的值（reflect value），即可以通过 rv.Field(si.index) 获取值
	in []string      // 嵌入结构体的父层结构体名称列表（忽略所有的 anonymous）；如果没有父层结构则是 nil
	si *sifterItem
}

func (cs *cachedSifter) SiftStruct(s interface{}, maxConfidentialLevel int) (map[string]interface{}, error) {
	// root reflect value (value of @param s)
	rrv := reflect.ValueOf(s)

	siList := make([]sifterItemCtx, 0, len(cs.sifterItems))
	for _, si := range cs.sifterItems {
		siList = append(siList, sifterItemCtx{
			rv: rrv,
			in: nil,
			si: si,
		})
	}

	// fmt.Printf("cachedSifter[%s]\n", cs)

	out := make(map[string]interface{}) // 最终的输出
	for idx := 0; idx < len(siList); idx++ {
		if idx > MAX_JSON_FIELD_NUMBER {
			return nil, fmt.Errorf("abort due to too many json fields (limit %d)", MAX_JSON_FIELD_NUMBER)
		}

		// current reflect value, parents' in list, sifter item
		curRv, curIn, curSi := siList[idx].rv, siList[idx].in, siList[idx].si

		if !curRv.Field(curSi.index).IsValid() || (curSi.isOmitEmpty && isEmptyValue(curRv.Field(curSi.index))) {
			continue
		}

		// 按照安全级别筛选域
		if curSi.cLevel > maxConfidentialLevel {
			continue
		}

		// fmt.Printf("curSi[%s]\n", curSi)

		if curSi.embedded == nil {
			// 处理非嵌入式结构体域的情况（按照 curIn 将值放在合适的节点上）
			end := out
			for _, parent := range curIn {
				if _, exist := end[parent]; !exist {
					end[parent] = make(map[string]interface{})
				}
				end = end[parent].(map[string]interface{})
			}
			end[curSi.alias] = curRv.Field(curSi.index).Interface()
		} else {
			// 处理嵌入式结构体的情况（将嵌入式域依次写FIFO等待处理）
			if !curSi.isAnonymous || (curSi.alias != "") {
				curIn = append(curIn, curSi.alias)
			}

			for _, si := range curSi.embedded.sifterItems {
				siList = append(siList, sifterItemCtx{
					rv: curRv.Field(curSi.index),
					in: curIn,
					si: si,
				})
			}
		}
	}
	return out, nil
}

func (cs *cachedSifter) String() string {
	var slist []string

	for _, si := range cs.sifterItems {
		if si.embedded != nil {
			// Note: 并非尾递归（golang 也不支持），注意调用栈嵌套问题；
			slist = append(slist, fmt.Sprintf("index[%d], field[%s], isAnonymous[%v], embedded[%s]",
				si.index, si.field, si.isAnonymous, si.embedded.String()))
		} else {
			slist = append(slist, fmt.Sprintf("index[%d], field[%s], alias[%s], isOmitEmpty[%v], cLevel[%d]",
				si.index, si.field, si.alias, si.isOmitEmpty, si.cLevel))
		}
	}

	return strings.Join(slist, "; ")
}

func (si *sifterItem) String() string {
	return fmt.Sprintf("index[%d], field[%s], alias[%s], isOmitEmpty[%v], cLevel[%d], isAnonymous[%v], hasEmbedded[%v]",
		si.index, si.field, si.alias, si.isOmitEmpty, si.cLevel, si.isAnonymous, si.embedded != nil)
}

// 针对某一个具体的结构体类型获取缓存的 sifter；如果不存在则将尝试新建对应的 sifter。
func GetSifter(rt reflect.Type) (cachedSifter, error) {
	sifterCache.RLock()
	cs, cached := sifterCache.m[rt]
	sifterCache.RUnlock()

	if cached {
		return cs, nil
	}

	cs, err := generateSifter(rt)
	if err != nil {
		return cachedSifter{}, err
	}

	sifterCache.Lock()
	if sifterCache.m == nil {
		sifterCache.m = make(map[reflect.Type]cachedSifter)
	}
	sifterCache.m[rt] = cs
	sifterCache.Unlock()

	return cs, nil
}

// 根据具体的结构体类型产生特定的 sifter。
func generateSifter(rt reflect.Type) (cachedSifter, error) {
	sList := make([]*sifterItem, 0)

	for i := 0; i < rt.NumField(); i++ {
		si := &sifterItem{
			index:       i,
			field:       rt.Field(i).Name,
			isAnonymous: rt.Field(i).Anonymous,
			// below are default values
			embedded:    nil,
			alias:       "",
			isOmitEmpty: false,
			cLevel:      CONFIDENTIAL_LEVEL0,
		}

		// 处理 json 标签
		if ignore, alias, omitempty, err := parseJsonTags(si.field, rt.Field(i).Tag.Get("json"), si.isAnonymous); err != nil {
			return cachedSifter{}, err
		} else if ignore {
			continue
		} else {
			si.alias = alias
			si.isOmitEmpty = omitempty
		}

		// 处理保密/脱敏标签
		if clevel, err := parseConfidentialTags(rt.Field(i).Tag.Get(TAG_CONFIDENTIAL)); err != nil {
			return cachedSifter{}, err
		} else {
			si.cLevel = clevel
		}

		// 处理嵌套的结构体
		if rt.Field(i).Type.Kind() == reflect.Struct {
			// embedded sifter
			eSifter, err := generateSifter(rt.Field(i).Type)
			if err != nil {
				return cachedSifter{}, err
			}
			si.embedded = &eSifter
		}

		sList = append(sList, si)
	}

	return cachedSifter{sifterItems: sList}, nil
}

// 可解析如下类型的 json 标签：
//
// 1. 没有标签
// 3. `json:"-"`
// 2. `json:"json_alias"`
// 4. `json:"json_alias,omitempty"`
// 4. `json:",omitempty"`
//
// @param
//  name - 结构体成员名称
//  jtag - json 标签字符串
//  isAnonymous - 是否是匿名域
// @return
//  ignore  - 是否在序列化过程中忽略此项成员（如：明确指定 `json:"-"` 或者首字母小写的非公开成员）
//  alias - 序列化时采用的别名，可能是 json 标签设置的别名，也可能是结构体域名称
//  omitempty - 当值为空时是否进行序列化
//
// Note:
//  1. 值为空的判定参考 json 标准库的文档：
//  false, 0, nil pointer or interface value, and any array, slice, map, or string of length zero
//  2. 仅支持匿名域类型是结构体的场景。
func parseJsonTags(name, jtag string, isAnonymous bool) (ignore bool, alias string, omitempty bool, err error) {
	// go struct field visibility
	if len(name) == 0 || unicode.IsLower(rune(name[0])) {
		ignore = true
		return
	}

	jtags := strings.Split(jtag, ",")
	switch len(jtags) {
	case 1:
		t := strings.TrimSpace(jtags[0])
		if t == "-" {
			// `json:"-"`
			ignore = true
			return
		} else if len(t) > 0 {
			// `json:"json_alias"`
			alias = t
			return
		} else {
			// 没有设置 json 标签
			if !isAnonymous {
				alias = name
			}
			return
		}
	case 2:
		t0, t1 := strings.TrimSpace(jtags[0]), strings.TrimSpace(jtags[1])
		if t1 != "omitempty" {
			ignore = true
			err = fmt.Errorf("json tag %s not supported", t1)
			return
		}
		if t0 == "" {
			// `json:",omitempty"`
			if !isAnonymous {
				alias = name
			}
			omitempty = true
			return
		} else {
			// `json:"json_alias,omitempty"`
			alias = t0
			omitempty = true
			return
		}
	default:
		// unsupported
		ignore = true
		err = fmt.Errorf("unsupported json tag string %s", jtag)
		return
	}
}

// 解析保密/脱敏相关的标签
//
//  1. 没有标签
//  2. `confidential:"-"`
//  3. `confidential:"level2"`
//
// @param
//  ctag - 保密级别/脱敏处理标签（confidential tag）
// @return
//  clevel - 保密级别（confidential level）
//
// Note:
//  如果没有标签或者是 `-`，那么都按照公开级别（即 level0）进行处理。
func parseConfidentialTags(ctags string) (clevel int, err error) {
	// default confidential level
	clevel = CONFIDENTIAL_LEVEL0

	clist := strings.Split(ctags, TAG_CONFIDENTIAL_SEPARATOR)
	switch len(clist) {
	case 1:
		clevelTag := strings.TrimSpace(clist[0])
		switch clevelTag {
		case "", CLEVEL_TAG_OMIT, CLEVEL_TAG_LEVEL0, CLEVEL_TAG_LEVEL1, CLEVEL_TAG_LEVEL2, CLEVEL_TAG_LEVEL3:
			clevel = _confidential_map[clevelTag]
			return
		default:
			err = fmt.Errorf("unsupported confidential level[%s]", clevelTag)
			return
		}
	default:
		err = fmt.Errorf("unsupported confidential tag[%s]", ctags)
		return
	}
}

// 判断某个值是否是空值（omitempty）
// @refer `/encoding/json/encode.go`
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
