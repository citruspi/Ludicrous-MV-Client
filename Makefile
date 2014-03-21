all:

	#clang lmv.c -o lmv -lcrypto -lssl -lcurl
	gcc -std=c99 -Wall lmv.c -o lmv -lcrypto -lssl -lcurl

go:

	go build lmv.go

clean:

	rm -f lmv

install:

	cp lmv /usr/local/bin/lmv

uninstall:

	rm -f /usr/local/bin/lmv
