all: clean lmv-client

lmv-client:

	go build lmv-client.go

clean:

	rm -f lmv-client

install:

	cp lmv /usr/local/bin/lmv-client

uninstall:

	rm -f /usr/local/bin/lmv-client
