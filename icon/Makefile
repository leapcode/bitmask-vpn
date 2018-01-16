.PHONY: all clean

all: on_unix.go off_unix.go wait_unix.go error_unix.go

%_unix.go: %.png
	echo "//+build linux darwin" > $@
	echo "" >> $@
	cat $< | 2goarray `echo ${<} | python -c "print raw_input().capitalize()[:-4]"` icon >> $@

clean:
	rm *_unix.go
