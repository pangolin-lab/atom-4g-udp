BINDIR=bin

#.PHONY: pbs

all: m a i test
#
#pbs:
#	cd pbs/ && $(MAKE)
#
test:
	go build  -ldflags '-w -s' -o $(BINDIR)/ctest mac/*.go
m:
	CGO_CFLAGS=-mmacosx-version-min=10.11 \
	CGO_LDFLAGS=-mmacosx-version-min=10.11 \
	GOARCH=amd64 GOOS=darwin go build  --buildmode=c-archive -o $(BINDIR)/dss.a mac/*.go
a:
	gomobile bind -v -o $(BINDIR)/dss.aar -target=android github.com/Iuduxras/atom-4g-udp/android
i:
	gomobile bind -v -o $(BINDIR)/iosLib.framework -target=ios github.com/Iuduxras/atom-4g-udp/ios

clean:
	gomobile clean
	rm $(BINDIR)/*