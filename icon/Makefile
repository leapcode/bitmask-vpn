.PHONY: all clean

all: 18_on_unix.go 18_off_unix.go 18_blocked_unix.go 18_wait_0_unix.go 18_wait_1_unix.go 18_wait_2_unix.go 18_wait_3_unix.go

go_array_name = "$(shell echo $(@) | sed -e 's/18_\(.*\)_unix\.go/\u\1/' -e 's/_//')"

18_%_unix.go: png/18/vpn_%.png
	echo "//+build linux darwin" > $@
	echo "" >> $@
	cat $< | 2goarray $(go_array_name) icon >> $@

clean:
	rm *_unix.go
