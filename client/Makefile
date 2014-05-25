all: lmv unlmv

lmv:

	go build client.go
	mv client lmv

unlmv:

	cp lmv unlmv

clean:

	rm -f lmv
	rm -f unlmv

install:

	cp lmv /usr/local/bin/lmv
	cp unlmv /usr/local/bin/unlmv

uninstall:

	rm -f /usr/local/bin/lmv
	rm -f /usr/local/bin/unlmv
