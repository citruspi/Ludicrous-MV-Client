all: clean lmv unlmv

lmv:

	go build lmv.go

unlmv:

	go build unlmv.go

clean:

	rm -f lmv unlmv

install:

	cp lmv /usr/local/bin/lmv
	cp unlmv /usr/local/bin/unlmv

uninstall:

	rm -f /usr/local/bin/lmv
	rm -f /usr/local/bin/unlmv
