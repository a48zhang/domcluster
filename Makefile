.PHONY: all clean

ROOT_DIR := $(shell pwd)

all:
	-mkdir -p $(ROOT_DIR)/built
	cd d8rctl && go build -o $(ROOT_DIR)/built/d8rctl main.go && cd ..
	cd domclusterd && go build -o $(ROOT_DIR)/built/domclusterd main.go && cd ..

clean:
	rm -r $(ROOT_DIR)/built