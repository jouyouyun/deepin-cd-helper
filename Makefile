CC = gcc
TARGET = deepin-cd-helper
DESTDIR = debian/tmp/
PREFIX = /usr

ifndef USE_GCCGO
	GOLDFLAGS = -ldflags '-s -w'
	GOBUILD = go build ${GOLDFLAGS}
else
	GOLDFLAGS = -s -w  -Os -O2
	GOLDFLAGS += $(shell pkg-config --libs gio-2.0)
	GOBUILD = go build -compiler gccgo -gccgoflags "${GOLDFLAGS}"
endif

all: ${TARGET}

${TARGET}:
	${GOBUILD} -o $@

install:
	mkdir -p ${DESTDIR}${PREFIX}/bin/
	cp -f ${TARGET} ${DESTDIR}${PREFIX}/bin/
	mkdir -p ${DESTDIR}${PREFIX}/share/dbus-1/system.d/
	cp -f data/dbus-1/system.d/com.deepin.helper.CD.conf ${DESTDIR}${PREFIX}/share/dbus-1/system.d/
	mkdir -p ${DESTDIR}${PREFIX}/share/dbus-1/system-services
	cp -f data/dbus-1/system-service/com.deepin.helper.CD.service ${DESTDIR}${PREFIX}/share/dbus-1/system-services
	mkdir -p ${DESTDIR}/lib/systemd/system/
	cp -f data/systemd/deepin-cd-helper.service ${DESTDIR}/lib/systemd/system/

clean:
	rm -f ${TARGET}

rebuild: clean ${TARGET}
