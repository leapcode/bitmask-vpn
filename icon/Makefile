.PHONY: all clean

unix := on_unix.go off_unix.go blocked_unix.go wait_0_unix.go wait_1_unix.go wait_2_unix.go wait_3_unix.go

darwin := on_darwin.go off_darwin.go blocked_darwin.go wait_0_darwin.go wait_1_darwin.go wait_2_darwin.go wait_3_darwin.go

all: $(unix) $(darwin)

go_array_name = "$(shell echo $(<) | sed -e 's/.*vpn_\(.*\)\.png/\u\1/' -e 's/_//')"

%_unix.go: png/white/vpn_%.png
	echo "//+build linux" > $@
	echo "// white icon"
	echo "" >> $@
	cat $< | 2goarray $(go_array_name) icon >> $@

%_darwin.go: png/black/vpn_%.png
	echo "//+build darwin" > $@
	echo "// black icon"
	echo "" >> $@
	cat $< | 2goarray $(go_array_name) icon >> $@

clean:
	rm *_unix.go
	rm *_darwin.go
