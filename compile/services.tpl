{{ range .Services }}
type {{.Name}}Service struct {
	ihttp.Service
	mux http.ServeMux
}
func New{{.Name}}Service(server *ihttp.Server) *{{.Name}}Service {
	var service {{.Name}}Service
	server.Mux.Handle("/{{.Name}}/", http.StripPrefix("/{{.Name}}", &service.mux))
	return &service
}
{{ $service := . }}
{{- range .Functions -}}
func (s *{{ $service.Name }}Service) {{.Name}}(f func({{ join .Inputs ", " }}) ({{ join .Outputs ", "}})) {
	s.mux.HandleFunc("/{{ .Name }}", func(w http.ResponseWriter, r *http.Request) {
		{{- if or (gt (len .Inputs) 0) (gt (len .Outputs) 0) }}
		var err error
		{{- end -}}
		{{- if gt (len .Inputs) 0 }}
		dec := gob.NewDecoder(r.Body)
		{{ end -}}
		{{- range $i, $e := .Inputs }}
		var in{{ $i }} {{ $e }}
		err = dec.Decode(&in{{ $i }})
		if err != nil {
			s.ErrorHandler(err)
			w.WriteHeader(500)
			return
		}
		{{- end }}

		{{- if gt (len .Outputs) 0 }}
		{{ arrToNames "out" .Outputs }} := f({{ arrToNames "in" .Inputs }})

		enc := gob.NewEncoder(w)
		{{- else }}
		f({{ arrToNames "in" .Inputs }})
		{{- end -}}
		{{ range $i, $el := .Outputs }}
		err = enc.Encode(&out{{ $i }})
		if err != nil {
			s.ErrorHandler(err)
			w.WriteHeader(500)
			return
		}
		{{- end }}
	})
}
{{ end }}
{{ end }}
