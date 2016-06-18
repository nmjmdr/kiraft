package raft


type InMemoryStable struct {
	m map[string]interface{}
}


func NewInMemoryStable() *InMemoryStable {
	i :=new(InMemoryStable)
	i.m = make(map[string]interface{})

	return i
}

func (i *InMemoryStable) GetUint64(key string) (uint64,bool) {

	v,ok := i.m[key]

		
	if !ok {
		return uint64(0),false
	}

	var u uint64
	u,ok = v.(uint64)
		
	if !ok {
		return uint64(0),false
	}

	return u,true
}


func (i *InMemoryStable) Get(key string) (string,bool) {

	v,ok := i.m[key]

	if !ok {
		return "",false
	}

	var u string
	u,ok = v.(string)

	if !ok {
		return "",false
	}

	return u,true
}


func (i *InMemoryStable) Store(key string,value interface{}) {
	i.m[key] = value
}
