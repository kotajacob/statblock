# statblock
# See LICENSE for copyright and license details.
.POSIX:

PREFIX ?= /usr/local
MANPREFIX ?= $(PREFIX)/share/man
GO ?= go
GOFLAGS ?=
RM ?= rm -f

all: statblock

statblock:
	$(GO) build $(GOFLAGS)

clean:
	$(RM) statblock

install: all
	mkdir -p $(DESTDIR)$(PREFIX)/bin
	cp -f statblock $(DESTDIR)$(PREFIX)/bin
	chmod 755 $(DESTDIR)$(PREFIX)/bin/statblock

uninstall:
	$(RM) $(DESTDIR)$(PREFIX)/bin/statblock

.DEFAULT_GOAL := all

.PHONY: all statblock clean install uninstall
