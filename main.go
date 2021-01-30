package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/pflag"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
)

var (
	asRaw   bool
	noTrunc bool
)

func main() {
	pf := pflag.NewFlagSet(`gJsonit`, pflag.ExitOnError)
	pf.BoolVarP(&asRaw, "raw", "r", false, "output raw")
	pf.BoolVarP(&noTrunc, "no-truc", "n", false, "no trunc")
	pf.Parse(os.Args[1:])
	args := pf.Args()
	var s string
	if len(args) > 0 {
		s = args[0]
	}
	switch {
	case StdinAvailable():
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			os.Exit(1)
		}
		if asRaw {
			gJsonIt(data, s)
			return
		}
		tableIt(data, s)
	default:
		fmt.Println("The command is intended to work with pipes.")
		fmt.Println("Usage:")
		fmt.Println("  cat/curl file.json/www... | gJsonIt")
		fmt.Println("  kubectl get ns -o json | gJsonIt")
	}
}

func StdinAvailable() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func gJsonIt(k []byte, s string) {
	var props gjson.Result
	if s != "" {
		props = gjson.GetBytes(k, s)
	} else {
		props = gjson.ParseBytes(k)
	}
	prettyPrint(props.String())
}

func tableIt(k []byte, s string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetReflowDuringAutoWrap(false)
	var sw bool
	var props gjson.Result
	if s != "" {
		props = gjson.GetBytes(k, s)
	} else {
		props = gjson.ParseBytes(k)
		table.SetHeader([]string{"❖", "▼"})
	}
	props.ForEach(func(key, value gjson.Result) bool {
		var val string
		switch {
		case noTrunc || len(value.String()) <= 120:
			val = value.String()
		case len(value.String()) > 120:
			val = truncateString(value.String(), 120)
		}
		td := []string{key.String(), val}
		table.Append(td)
		if key.String() != "" {
			sw = true
		} else {
			sw = false
		}
		return true
	})
	if sw == true {
		table.SetHeader([]string{"︎︎︎▲ " + s + "︎ ▼", ""})
	} else {
		table.SetHeader([]string{"︎︎︎▲ " + s + "︎ ►", "▼"})
	}
	table.Render()
}

func truncateString(str string, num int) string {
	str = str[0:num] + "..."
	return str
}

func prettyPrint(json interface{}) {
	switch json := json.(type) {
	case []byte:
		fmt.Printf("%s", pretty.Pretty(json))
	case string:
		j := []byte(json)
		fmt.Printf("%s", pretty.Pretty(j))
	}
}
