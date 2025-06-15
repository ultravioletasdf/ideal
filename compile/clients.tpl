{{ range .Services }}
type {{.Name}}Client struct {
	*ihttp.Client
}
func New{{.Name}}Client(client *ihttp.Client) (*{{.Name}}Client) {
	return &{{.Name}}Client{Client: client}
}
{{- $service := . -}}
{{ range .Functions}}
func (c *{{ $service.Name }}Client) {{.Name}}({{ arrToArguments "in" .Inputs false }}) ({{ arrToArguments "out" .Outputs true }}) {
	body := new (bytes.Buffer)
	{{ if gt (len .Inputs) 0 -}}
	enc := gob.NewEncoder(body)
	{{ end }}
	{{- range $i, $el := .Inputs -}}
	enc.Encode(&in{{ $i }})
	{{ end -}}

	{{ if gt (len .Outputs) 0 }}
	var res io.ReadCloser
	res, err = c.Call("/{{ $service.Name }}/{{ .Name }}", body)
	dec := gob.NewDecoder(res)
	{{- else }}
	_, err = c.Call("/{{ $service.Name }}/{{ .Name }}", body)
	{{- end }}
	if err != nil {
		return
	}
	{{ range $i, $el := .Outputs -}}
	dec.Decode(&out{{ $i }})
	{{ end }}
	return
}
{{ end }}
{{ end }}
