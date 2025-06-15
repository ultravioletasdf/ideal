{{ range .Structures }}
type {{ .Name }} struct {
{{- range .Fields -}}
	{{- $type := .Type -}}
	{{- if eq .Type "float" -}}
		{{ $type = "float64" }}
	{{- else if eq .Type "int" -}}
		{{ $type = "int64" }}
	{{- end }}
	{{.Name}} {{ .Type }}
{{- end }}
}
func (d *{{ .Name }}) Encode() (bin []byte, err error) {
	buf := new (bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err = enc.Encode(d)
	if err != nil {
		return
	}
	bin = buf.Bytes()
	return
}
func (d *{{ .Name }}) Decode(bin []byte) error {
	buf := bytes.NewBuffer(bin)
	dec := gob.NewDecoder(buf)
	return dec.Decode(d)
}
{{ end }}
