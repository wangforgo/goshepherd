package main

import (
	"testing"
)

func TestShepherdSheepManagement(t *testing.T) {
	s := newShepherd()

	checkSheep(t, s.dumpSheep(), nil)

	s.addSheep(&Sheep{port: 8081})
	s.addSheep(&Sheep{port: 8082})
	s.addSheep(&Sheep{port: 8083})
	checkSheep(t, s.dumpSheep(), []int{8081,8082,8083})

	s.rmvSheep(8082)
	checkSheep(t, s.dumpSheep(), []int{8081,8083})

	s.rmvSheep(8085)
	s.rmvSheep(8081)
	s.rmvSheep(8083)
	checkSheep(t, s.dumpSheep(), nil)

	s.addSheep(&Sheep{port: 8086})
	s.addSheep(&Sheep{port: 8087})
	s.addSheep(&Sheep{port: 8088})
	checkSheep(t, s.dumpSheep(), []int{8086,8087,8088})

}


func checkSheep(t *testing.T, allSheep []*Sheep, expect []int) {
	if len(allSheep) != len(expect) {
		t.Fail()
	}
	for i:= range allSheep {
		if allSheep[i].port != expect[i] {
			t.Fail()
		}
	}
}