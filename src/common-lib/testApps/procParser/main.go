package main

import (
	"fmt"
	"strconv"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/env"
	pp "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/procParser"
)

func main() {
	factory := new(pp.ParserFactoryImpl)
	parser := factory.GetParser()
	cfg := pp.Config{}
	cfg.ParserMode = pp.ModeTabular
	//cfg.KeyField = 3

	reader, err := env.FactoryEnvImpl{}.GetEnv().GetCommandReader("ps", "-f", "-p", strconv.Itoa(27729), "-eo", "pid,pcpu,nlwp,pmem,state,start_time")
	if err != nil {
		fmt.Println(err)
	}

	defer reader.Close()

	data, err := parser.Parse(cfg, reader)
	/*file, err := os.Open("/proc/meminfo")*/
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(data)
	// fmt.Println(data.Lines[0].Values[0])
	// fmt.Println(data.Lines[1].Values[2])
	// fmt.Println(data.Lines[7].Values[2])

	/*scanner := bufio.NewScanner(file)
	i := 0
	for scanner.Scan() {
		fmt.Println(i)
		fmt.Println(scanner.Text())
		i++
	}*/
}
