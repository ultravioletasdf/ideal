package {{ .Package }}
import (
    {{ if or (gt (len .Structures) 0) (gt (len .Services) 0) -}}
	"bytes"
	"encoding/gob"
	{{- end -}}
	{{ if gt (len .Services) 0 }}
	"net/http"
	ihttp "github.com/ultravioletasdf/ideal/http"
	"io"
	{{- end -}}
)

{{template "structs.tpl" .}}
{{template "services.tpl" .}}
{{template "clients.tpl" .}}
