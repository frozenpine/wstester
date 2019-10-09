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
	formatStr string
)

func init() {
	flag.StringVar(&formatStr, "format", "",
		"Go template string for output.")
}

func makeTemplate() (tpl *template.Template) {
	var err error

	if formatStr != "" {
		tpl, err = template.New("formater").Parse(formatStr)
		if err != nil {
			panic(err)
		}
	}

	return
}

func format(ctx context.Context, table string, tpl *template.Template, ch <-chan models.TableResponse) {
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
