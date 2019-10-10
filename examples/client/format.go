package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"text/template"

	"github.com/frozenpine/wstester/models"
	flag "github.com/spf13/pflag"
)

var (
	formatList []string
)

func init() {
	flag.StringSliceVar(&formatList, "format", []string{},
		"Go template string for output, mutiple format must have same order with --output.")
}

func makeTemplates() map[string]*template.Template {
	tableLen := len(tables)

	templates := make(map[string]*template.Template)

	for idx, format := range formatList {
		tpl, err := template.New("formater").Parse(format)

		if err != nil {
			panic(err)
		}

		if idx >= tableLen {
			break
		}

		templates[tables[idx]] = tpl
	}

	return templates
}

func format(ctx context.Context, table string, templates map[string]*template.Template, ch <-chan models.TableResponse) {
	for {
		select {
		case <-ctx.Done():
			return
		case rsp := <-ch:
			if rsp == nil {
				return
			}

			action := rsp.GetAction()

			for _, data := range rsp.GetData() {
				var result string

				if tpl, exist := templates[table]; exist {
					buf := bytes.Buffer{}

					if err := tpl.Execute(&buf, data); err != nil {
						panic(err)
					}

					result = buf.String()
				} else {
					d, _ := json.Marshal(data)

					result = string(d)
				}

				log.Println(table, action, "<-", result)
			}
		}
	}
}
