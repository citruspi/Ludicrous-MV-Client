all: clean lmv

lmv:

	go build lmv.go

clean:

	rm -f lmv

install:

	cp lmv /usr/local/bin/lmv

uninstall:

	rm -f /usr/local/bin/lmv
