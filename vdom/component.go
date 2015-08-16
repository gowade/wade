package vdom

type Component interface {
	VDOMRender() *Element
}

type RenderData struct {
	vdomIndex int
	Children  []Node
	Refs      interface{}
}

var (
	mComponentData = make(map[Component]*RenderData)
)

func GetComponentData(com Component) *RenderData {
	return mComponentData[com]
}

func CreateComponentData(com, old Component) *RenderData {
	cdata := mComponentData[com]
	if com == nil {
		if old != nil {
			delete(mComponentData, old)
		}

		cdata = &RenderData{}
		mComponentData[com] = cdata
	}

	return cdata
}

func Rerender(component Component) {
}
