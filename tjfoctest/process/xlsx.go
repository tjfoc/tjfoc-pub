package process

import (
	"fmt"
	"os"

	"github.com/360EntSecGroup-Skylar/excelize"
)

const (
	NAME      = "Sheet1"
	FUNCTION1 = "NewTx"
	FUNCTION2 = "GetHeight"
	FUNCTION3 = "GetBlockByHash"
	FUNCTION4 = "GetBlockByHeight"
	FUNCTION5 = "GetTx"
	FUNCTION6 = "GetTxBlock"
	SAVENAME  = "./testData.xlsx"
)

type TestData struct {
	Function               string
	RutineTimes            int
	CreateConn             float64
	RequestAndResponseTime float64
	Result                 string
}

func xlsxInit() *xlsx {
	var x *excelize.File
	if _, err := os.Stat("./testData.xlsx"); os.IsNotExist(err) {
		x = excelize.NewFile()
	} else {
		x, err = excelize.OpenFile("./testData.xlsx")
		if err != nil {
			panic(err)
		}
	}

	x.SetCellValue("Sheet1", "A1", "功能")
	x.SetCellValue("Sheet1", "B1", "并发次数")
	x.SetCellValue("Sheet1", "C1", "链接时间")
	x.SetCellValue("Sheet1", "D1", "请求回复时间")
	x.SetCellValue("Sheet1", "E1", "结果")

	return &xlsx{
		file:  x,
		index: getIndex(x),
	}
}

type xlsx struct {
	file  *excelize.File
	index int
}

func (x *xlsx) addFunction(inData *TestData) {
	x.file.SetCellValue(NAME, x.format("A"), inData.Function)
	x.file.SetCellValue(NAME, x.format("B"), inData.RutineTimes)
	x.file.SetCellValue(NAME, x.format("C"), inData.CreateConn)
	x.file.SetCellValue(NAME, x.format("D"), inData.RequestAndResponseTime)
	x.file.SetCellValue(NAME, x.format("F"), inData.Result)
	x.index += 1
}

/*func (x *xlsx) addBatchFunction(indata []TestData) {
	length := len(indata)
	for i := 0; i < length; i++ {
		x.addFunction(indata[i])
	}
}*/

func (x *xlsx) format(a string) string {
	return fmt.Sprintf("%s%d", a, x.index)
}

func (x *xlsx) Save() {
	err := x.file.SaveAs(SAVENAME)
	if err != nil {
		panic(err)
	}
}

func getIndex(f *excelize.File) int {
	for i := 1; ; i++ {
		s := f.GetCellValue(NAME, fmt.Sprintf("%s%d", "A", i))
		if s == "" {
			return i
		}
	}
}
