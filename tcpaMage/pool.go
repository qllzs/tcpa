package main

import (
	"net/rpc"
	"sync"
)

type state struct {
	cli    *rpc.Client
	IsIdle bool
}

var taPool map[string]*state
var taMux sync.Mutex

var xgwPool sync.Map //key: ueIP   value:taIP

func init() {
	taPool = make(map[string]*state)
}

func addIPToTaPool(taIP string) {

	taMux.Lock()
	defer taMux.Unlock()
	if taPool[taIP] != nil {
		delete(taPool, taIP)
	}

	taPool[taIP] = &state{IsIdle: true}
}

func getTaFromTaPool(taIP string) *state {
	taMux.Lock()
	defer taMux.Unlock()

	return taPool[taIP]

}

func getTaFromXgwPool(ueIP string) string {

	tcpaIP, ok := xgwPool.Load(ueIP)

	if ok {
		return tcpaIP.(string)
	}

	return ""
}

func getFreeIPFromTaPool() string {

	taMux.Lock()
	defer taMux.Unlock()

	for k, v := range taPool {
		if v.IsIdle == true {
			v.IsIdle = false
			return k
		}
	}

	return ""
}

func deleteIPFromTaPool(taIP string) {
	taMux.Lock()
	defer taMux.Unlock()

	delete(taPool, taIP)
}

func storeToXGWPool(ueIP string, taIP string) {

	xgwPool.Store(ueIP, taIP)
}

func loadFromXGWPool(ueIP string) (string, bool) {

	taIP, ok := xgwPool.Load(ueIP)
	if taIP == nil {
		return "", ok
	}
	return taIP.(string), ok
}

func deleteIPFromXGWPoolByTaIP(tcpaIP string) {

	xgwPool.Range(func(k, v interface{}) bool {
		if v == tcpaIP {
			xgwPool.Delete(k)
		}
		return true
	})
}
