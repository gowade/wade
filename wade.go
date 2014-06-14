package wade

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	jq "github.com/gopherjs/jquery"
)

var (
	gRivets  = js.Global.Get("rivets")
	gHistory = js.Global.Get("history")
	gJQ      = jq.NewJQuery
)

const (
	AttrPrefix = "attr-"
	BindPrefix = "bind-"
)

type Wade struct {
	pm      *PageManager
	custags map[string]interface{} //custom tags
}

type ErrorMap map[string]map[string]interface{}

type Validated struct {
	Errors ErrorMap
}

type ErrorsBinding struct {
	Errors *ErrorMap
}

func (v *Validated) Init(dataModel interface{}) {
	m := make(ErrorMap)
	typ := reflect.TypeOf(dataModel)
	if typ.Kind() != reflect.Struct {
		panic("Validated data model passed to Init() must be a struct.")
	}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		m[f.Name] = make(map[string]interface{})
	}
	v.Errors = m
}

type pageInfo struct {
	path  string
	title string
}

type PageHandler func() interface{}

type PageManager struct {
	router       js.Object
	currentPage  *jq.JQuery
	pageHandlers map[string][]PageHandler
	startPage    string
	basePath     string
	pages        map[string]pageInfo
	notFoundPage string
	pageModels   []js.Object
}

func WadeUp(startPage, basePath string) *Wade {
	return &Wade{
		pm:      newPageManager(startPage, basePath),
		custags: make(map[string]interface{}),
	}
}

func newPageManager(startPage, basePath string) *PageManager {
	return &PageManager{
		router:       js.Global.Get("RouteRecognizer").New(),
		currentPage:  nil,
		pageHandlers: make(map[string][]PageHandler),
		basePath:     basePath,
		startPage:    startPage,
		pages:        make(map[string]pageInfo),
		notFoundPage: "",
		pageModels:   make([]js.Object, 0),
	}
}

func rivetsConf() {
	//Rivets configs here
}

func (wd *Wade) run() {
	rivetsConf()
}

func (pm *PageManager) cutPath(path string) string {
	if strings.HasPrefix(path, pm.basePath) {
		path = path[len(pm.basePath):]
	}
	return path
}

func (pm *PageManager) page(pageId string) pageInfo {
	if page, ok := pm.pages[pageId]; ok {
		return page
	}
	panic(fmt.Sprintf("no such page #%v found.", pageId))
}

func (pm *PageManager) SetNotFoundPage(pageId string) {
	_ = pm.page(pageId)
	pm.notFoundPage = pageId
}

func (pm *PageManager) Url(path string) string {
	return pm.basePath + path
}

func documentUrl() string {
	location := gHistory.Get("location")
	if location.IsNull() || location.IsUndefined() {
		location = js.Global.Get("document").Get("location")
	}
	return location.Get("pathname").Str()
}

func (pm *PageManager) setupPageOnLoad() {
	path := pm.cutPath(documentUrl())
	if path == "/" {
		path = pm.page(pm.startPage).path
		gHistory.Call("replaceState", nil, pm.pages[pm.startPage].title, pm.Url(path))
	}
	pm.updatePage(path)
}

func (pm *PageManager) getReady() {
	gJQ("a").On(jq.CLICK, func(e jq.Event) {
		e.PreventDefault()
		a := gJQ(e.Target)
		href := a.Attr("href")
		if href == "" || !strings.HasPrefix(href, ":") {
			return
		}

		pageId := string([]rune(href)[1:])
		pageInf := pm.page(pageId)
		gHistory.Call("pushState", nil, pageInf.title, pm.Url(pageInf.path))
		pm.updatePage(pageInf.path)
	})

	gJQ(js.Global.Get("window")).On("popstate", func() {
		pm.updatePage(documentUrl())
	})

	gJQ("welement").Hide()
	pm.setupPageOnLoad()
}

func pageHide(elems jq.JQuery) {
	elems.Hide()
	elems.SetData("hidden", "t")
}

func pageShow(elems jq.JQuery) {
	elems.Show()
	elems.SetData("hidden", "")
}

func pageIsHidden(pageElem jq.JQuery) bool {
	return pageElem.Data("hidden") == "t"
}

func (pm *PageManager) updatePage(url string) {
	url = pm.cutPath(url)
	matches := pm.router.Call("recognize", url)
	println("path: " + url)
	if matches.IsUndefined() || matches.Length() == 0 {
		if pm.notFoundPage != "" {
			pm.updatePage(pm.page(pm.notFoundPage).path)
		} else {
			panic("Page not found. No 404 handler declared.")
		}
	}

	pageId := matches.Index(0).Get("handler").Invoke().Str()
	pageElem := gJQ("#" + pageId)
	gJQ("title").SetText(pm.page(pageId).title)
	if pm.currentPage != nil {
		cp := pm.currentPage
		pageHide(*cp)
		cp.Parents("wpage").Each(func(idx int, p jq.JQuery) {
			pageHide(p)
		})
	}

	pageElem.Parents("wpage").Each(func(idx int, p jq.JQuery) {
		pageShow(p)
	})
	pageShow(pageElem)
	pm.currentPage = &pageElem
	if handlers, ok := pm.pageHandlers[pageId]; ok {
		for _, handler := range handlers {
			model := handler()
			modelo := gRivets.Call("bind", pageElem.Underlying(), model).Get("models")
			//pm.pageModels = make([]js.Object, nmodels)
			//for i := 0; i < nmodels; i++ {
			//	pm.pageModels[i] = models.Index(i)
			//}
			pm.pageModels = append(pm.pageModels, modelo)
		}
	}
}

func (pm *PageManager) inCurrentPage(elem jq.JQuery) bool {
	//return elem.Closest(pm.currentPage).Length > 0
	return !pageIsHidden(elem.Parent("wpage").First())
}

func (pm *PageManager) RegisterHandler(pageId string, handlerFn PageHandler) {
	if _, exist := pm.pageHandlers[pageId]; !exist {
		pm.pageHandlers[pageId] = make([]PageHandler, 0)
	}
	pm.pageHandlers[pageId] = append(pm.pageHandlers[pageId], handlerFn)
}

func (pm *PageManager) RegisterPages(pages map[string]string) {
	for path, pageId := range pages {
		if _, exist := pm.pages[pageId]; exist {
			panic(fmt.Sprintf("Page #%v has already been registered.", pageId))
		}
		pageElem := gJQ("#" + pageId)
		if pageElem.Length == 0 {
			panic(fmt.Sprintf("There is no such page element #%v.", pageId))
		}

		(func(path, pageId string) {
			pm.router.Call("add", []map[string]interface{}{
				map[string]interface{}{
					"path": path,
					"handler": func() string {
						return pageId
					},
				},
			})
		})(path, pageId)

		pm.pages[pageId] = pageInfo{path: path, title: pageElem.Attr("title")}
	}
}

func (wd *Wade) RegisterNewTag(tagid string, model interface{}) {
	tagElem := gJQ("#" + tagid)
	if tagElem.Length == 0 {
		panic(fmt.Sprintf("Welement with id #%v does not exist.", tagid))
	}
	if tagElem.Prop("tagName") != "WELEMENT" {
		panic(fmt.Sprintf("The element #%v to register new tag must be a welement.", tagid))
	}
	wd.custags[tagid] = model
}

func (wd *Wade) Pager() *PageManager {
	return wd.pm
}

func (wd *Wade) Start() {
	gJQ(js.Global.Get("document")).Ready(func() {
		pageHide(gJQ("wpage"))
		wd.pm.getReady()
		for tagid, model := range wd.custags {
			mtype := reflect.TypeOf(model)
			if mtype.Kind() != reflect.Struct {
				panic(fmt.Sprintf("Wrong type for the model of tag #%v, it must be a struct (non-pointer).", tagid))
			}
			wd.bind(tagid, mtype)
		}
	})
}

func (wd *Wade) bind(tagid string, model reflect.Type) {
	tagElem := gJQ("#" + tagid)
	publicAttrs := []string{}
	if attrs := tagElem.Attr("attributes"); attrs != "" {
		publicAttrs = strings.Split(attrs, " ")
		for _, attr := range publicAttrs {
			if _, ok := model.FieldByName(attr); !ok {
				panic(fmt.Sprintf(`Attribute "%v" is not available in the model for custom tag "%v".`, attr, tagid))
			}
		}
	}

	bindables := []string{}
	if bdbs := tagElem.Attr("bindables"); bdbs != "" {
		bindables = strings.Split(bdbs, " ")
		for _, bdb := range bindables {
			if _, ok := model.FieldByName(bdb); !ok {
				panic(fmt.Sprintf(`Bindable "%v" is not available in the model for custom tag "%v".`, bdb, tagid))
			}
		}
	}

	elems := gJQ(tagid)
	elems.Each(func(idx int, elem jq.JQuery) {
		elem.Append(tagElem.Html())
	})
	elems.Each(func(idx int, elem jq.JQuery) {
		if !wd.pm.inCurrentPage(elem) {
			return
		}
		cptr := reflect.New(model)
		clone := cptr.Elem()
		for _, attr := range publicAttrs {
			if val := elem.Attr(AttrPrefix + attr); val != "" {
				field := clone.FieldByName(attr)
				var err error = nil
				var v interface{}
				ftype, _ := model.FieldByName(attr)
				switch ftype.Type.Kind() {
				case reflect.Int:
					v, err = strconv.Atoi(val)
				case reflect.Uint:
					var m uint32
					var n uint64
					n, err = strconv.ParseUint(val, 10, 32)
					m = uint32(n)
					v = m
				case reflect.Float32:
					v, err = strconv.ParseFloat(val, 32)
				case reflect.Bool:
					v, err = strconv.ParseBool(val)
				case reflect.String:
					v = val
				default:
					err = fmt.Errorf(`Unhandled type for attribute "%v" of custom tag "%v".`, attr, tagid)
				}

				if err != nil {
					panic(fmt.Sprintf(`Invalid value "%v" for attribute "%v" of custom tag "%v": type mismatch. Parse info: %v.`,
						val, attr, tagid, err))
				}

				field.Set(reflect.ValueOf(v))
			}
		}

		for _, bdb := range bindables {
			if target := elem.Attr(BindPrefix + bdb); target != "" {
				ok := false
				for _, pgmodel := range wd.pm.pageModels {
					pgm := reflect.ValueOf(pgmodel.Interface())
					if pgm.Kind() == reflect.Ptr {
						pgm = pgm.Elem()
					}
					m := pgm.FieldByName(target)
					if m.IsValid() {
						f := clone.FieldByName(bdb)
						if f.Kind() != reflect.Ptr {
							panic(fmt.Sprintf(`Field "%v" of custom tag "%v" is not bindable because it's not a pointer.`, tagid, bdb))
						}
						f.Set(m.Addr())
						//println(f.Interface())
						ok = true
						break
					}
				}
				if !ok {
					panic(fmt.Sprintf(`binding failed for bindable "%v" of custom tag "%v",
					the bind target "%v" is not available.`, bdb, tagid, target))
				}
			}
		}

		gRivets.Call("bind", elem.Underlying(), cptr.Interface())
	})
}
