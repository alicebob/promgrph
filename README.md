Embedabble (mostly) prometheus/victoria server side graphs.

Goal:
 - server connects to prometheus
 - embeddable with Go templates
 - secure (as in the prom expression never ends up client side)
 - JS is fine, but not required


## Status

Proof of concept. You can get some basic graphs, and the template idea works.


## AI


Big picture is old fashioned manual work, but Mistral filled in details (esp the svg rendering).
