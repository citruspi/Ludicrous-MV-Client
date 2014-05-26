all: clean server

clean:

	rm -f lmv-server

server:

	go build server.go
	mv server lmv-server

install:

	mv lmv-server /usr/local/bin/.

uninstall:

	rm -f /usr/local/bin/lmv-server
