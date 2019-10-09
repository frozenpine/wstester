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
	format string
)

func init() {
	flag.StringVar(&format, "format", "",
		"Go template string for output.")
}

func formatTemplate() (tpl *template.Template) {
	var err error

	if format != "" {
		tpl, err = template.New("formater").Parse(format)
		if err != nil {
			panic(err)
		}
	}

	return
}

func output(ctx context.Context, table string, tpl *template.Template, ch <-chan models.TableResponse) {
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
				if tpl != nil {
					buf := bytes.Buffer{}

					if err := tpl.Execute(&buf, data); err != nil {
						panic(err)
					}

					result = buf.String()
				} else {
					d, _ := json.Marshal(data)

					result = string(d)
				}

				log.Println(table, action, ":", result)
			}
		}
	}
}
