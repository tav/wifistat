build: analysis.js site.css
	@rm -f index.html
	@cat template-head.html >> index.html
	@echo "<script>" >> index.html
	@uglifyjs analysis.js >> index.html
	@echo "</script><style>" >> index.html
	@cat site.css >> index.html
	@echo "</style>" >> index.html
	@cat template-tail.html >> index.html

site.css: site.sass
	@sass site.sass site.css -t compressed

analysis.js: analysis.coffee
	@coffee -c analysis.coffee
