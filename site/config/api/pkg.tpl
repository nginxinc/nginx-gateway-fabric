{{ define "packages" }}
---
title: "API Reference"
description: "NGINX Gateway API Reference"
weight: 100
toc: true
---

{{ with .packages}}
<p>Packages:</p>
<ul>
    {{ range . }}
    <li>
        <a href="#{{- packageAnchorID . -}}">{{ packageDisplayName . }}</a>
    </li>
    {{ end }}
</ul>
{{ end}}

{{ range .packages }}
    <h2 id="{{- packageAnchorID . -}}">
        {{- packageDisplayName . -}}
    </h2>

    {{ with (index .GoPackages 0 )}}
        {{ with .DocComments }}
        <div>
            {{ safe (renderComments .) }}
        </div>
        {{ end }}
    {{ end }}

    {{- if (gt 0 (len (visibleTypes (sortedTypes .Types)))) -}}
    Resource Types:
    {{- end -}}
    <ul>
    {{- range (visibleTypes (sortedTypes .Types)) -}}
        {{ if isExportedType . -}}
        <li>
            <a href="{{ linkForType . }}">{{ typeDisplayName . }}</a>
        </li>
        {{- end }}
    {{- end -}}
    </ul>

    {{ range (visibleTypes (sortedTypes .Types))}}
        {{ template "type" .  }}
    {{ end }}
    <hr/>
{{ end }}

<p><em>
    Generated with <code>gen-crd-api-reference-docs</code>
</em></p>

{{ end }}
