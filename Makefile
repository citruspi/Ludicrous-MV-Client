all: clean unlmv

unlmv:

	go build unlmv.go

clean:

	rm -f unlmv

install:

	cp unlmv /usr/local/bin/unlmv

uninstall:

	rm -f /usr/local/bin/unlmv
