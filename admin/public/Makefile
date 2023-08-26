.PHONY: all

all: *.php
	@err=0; for file in $^ ; do php -l $$file >/dev/null || { echo "Syntax error in $$file" ; err=1; } ; done ;\
	[ $$err -eq 0 ] && echo "No syntax errors"
