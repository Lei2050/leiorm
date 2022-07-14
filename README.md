# LEIORM

The ORM tool for redis developed with golang

## Feature & Example usage

```golang
type Attr struct {
	Type1 int64 `leiorm:"type1"`
	Val1  int64 `leiorm:"val1"`
	Type2 int64 `leiorm:"type2"`
	Val2  int64 `leiorm:"val2"`
}

type Item struct {
	Id   int   `leiormpri:"id"`
	Type int64 `leiorm:"type"`
	Size int64 `leiorm:"size"`
}

type Item2 struct {
	Type int64 `leiorm:"type"`
	Size int64 `leiorm:"size"`
}

type Equip struct {
	Id    uint64 `leiormpri:"id"`
	Attr  `leiorm:"attr"`
	Level uint64 `leiorm:"lv"`
}

type Knight struct {
	Fuk   string `leiormpri:"id"`
	Level int    `leiorm:"lv"`
	Star  int    `leiorm:"star"`
}

type Role struct {
	/* for example, Role.Id = 1001 */

	/*
		HMSET Role:1001
			id 	 	  1001
			ass 	  e;e;..
			arr2d     e-e..;e-e..;..
			attrs     k:e|k:e|..
			attrs2    k:e|k:e|..
			maparr    k:e;e;..|k:e;e;..|...
	*/
	Id     uint64              `leiormpri:"id"`
	Ass    []int64             `leiorm:"ass"`
	Arr2d  [][]float64         `leiorm:"arr2d"`
	Attrs  map[int]uint64      `leiorm:"attrs"`
	Attrs2 map[int]uint64      `leiorm:"attrs2"`
	Maparr map[uint64][]string `leiorm:"maparr"`

	/*
		SADD Role:1001:itemss Items[0].id Items[1].id ...
		HMSET Role:1001:items:{Items[0].id}
			id 		...
			type    ...
			size    ...
		HMSET Role:1001:items:{Items[1].id}
			id 		...
			type    ...
			size    ...
		...
	*/
	Items []*Item `leiorm:"items"`

	/*
		If tag 'leiormpri' is not specified, then auto incrementing id is used.

		SADD Role:1001:itemss 1 2 ...
		HMSET Role:1001:items:1
			type    ...
			size    ...
		HMSET Role:1001:items:2
			type    ...
			size    ...
	*/
	Items2 []*Item2 `leiorm:"items2"`

	/*
		HMSET role:1001:equip
			id    ...
			lv    ...
		HMSET role:1001:equip:attr
			type1    ...
			val1     ...
			type2    ...
			val2     ...
	*/
	Equip  *Equip `leiorm:"equip"`
	Equip2 *Equip `leiorm:"eq"`

	/*
		SADD Role:1001:knightss k1 k2 ...
		HMSET Role:1001:knights:{k1}
			id    ...
			lv    ...
			star  ...
		HMSET Role:1001:knights:{k2}
			id    ...
			lv    ...
			star  ...
		...
	*/
	Knights map[string]*Knight `leiorm:"knights"`
	Ks      map[string]*Knight `leiorm:"ks"`
}

func TestLoad() {
	rd, _ := redis.DialURL(url)

	var role = Role{
		Id:     10010321,
		Ass:    []int64{11, 22, 33},
		Items:  []*Item{{44, 10001, 1}, {55, 10002, 2}},
		Items2: []*Item2{{10301, 1}, {10302, 2}},
		Attrs: map[int]uint64{
			101: 30,
			202: 100,
			303: 66,
		},
		Arr2d: [][]float64{{1.1, 2.2, 3.3}, {4.4, 5.5, 6.6}, {7.7, 8.8, 9.9}},
		Attrs2: map[int]uint64{
			111: 40,
			212: 110,
			313: 76,
		},
		Equip: &Equip{
			Id:    8001001,
			Attr:  Attr{77, 20, 88, 30},
			Level: 30,
		},
		Equip2: &Equip{
			Id:    8001002,
			Attr:  Attr{80, 30, 90, 50},
			Level: 50,
		},
		Maparr: map[uint64][]string{
			666: {"hello", "world"},
			999: {"f", "y"},
		},
		Knights: map[string]*Knight{
			"20001": {"20001", 34, 1},
			"20002": {"20002", 50, 5},
		},
		Ks: map[string]*Knight{
			"30001": {"30001", 34, 1},
			"30002": {"30002", 50, 5},
		},
	}
	role.Id = 1001
	SaveModel(rd, &role, nil)
	SaveModel(rd, true, "testbooltrue")
	SaveModel(rd, false, "testboolfalse")
	SaveModel(rd, "hello", "teststr")
	SaveModel(rd, 3.14, "testpi")
	SaveModel(rd, []int{1, 2, 3}, "testslc")
	SaveModel(rd, [2]int{984, 472}, "testarr")
	SaveModel(rd, [][]int{{1, 1}, {2, 2}, {3, 3}}, "testslc2d")
	SaveModel(rd, [][2]int{{1, 1}, {2, 2}, {3, 3}}, "testslcarr")
	SaveModel(rd, [2][]int{{1, 1, 1}, {2, 2, 2}}, "testarrslc")
	SaveModel(rd, [2][3]int{{1, 1, 1}, {2, 2, 2}}, "testarrarr")
	SaveModel(rd, map[int]string{1: "hello", 2: "world"}, "testmap")
	SaveModel(rd, map[int][]int{1: {333, 444}, 2: {555, 666}}, "testmaparr")
}

func TestLoad() {
	rd, _ := redis.DialURL(url)

	r := &Role{}
	LoadModel(rd, r, "Role:1001")

	var b bool
	LoadModel(rd, &b, "testbooltrue")
	LoadModel(rd, &b, "testboolfalse")

	var s string
	LoadModel(rd, &s, "teststr")

	var f float32
	LoadModel(rd, &f, "testpi")

	var slc []int
	LoadModel(rd, &slc, "testslc")

	var arr [2]int
	LoadModel(rd, &arr, "testarr")

	var slcarr [][2]int
	LoadModel(rd, &slcarr, "testslcarr")

	var arrslc [2][]int
	LoadModel(rd, &arrslc, "testarrslc")

	var arrarr [2][3]int
	LoadModel(rd, &arrarr, "testarrarr")

	var m map[int]string
	LoadModel(rd, &m, "testmap")

	var maparr map[int][]int
	LoadModel(rd, &maparr, "testmaparr")
}
```

## Unsupported type
* Arrays or slices with more than 2 dimensions
* Multilevel pointer
* Complex nested types, such as map[&struct]map[int]struct

Note that, trying to use unsupported type will lead to undefined results.