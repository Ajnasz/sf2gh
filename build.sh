#!/bin/sh

build() {
	VERSION=$(git describe --tags)
	BUILD=$(date +%FT%T%z)
	ARCHLIST="$1"
	OSLIST="$2"
	echo $VERSION
	echo $BUILD
	for os in $OSLIST
	do
		for arch in $ARCHLIST
		do
			echo "building $os.$arch"
			GOOS=$os GOARCH=$arch go build -ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD}" -o "sf2gh.$os.$arch"
		done
	done
}

remove() {
	for os in $OSLIST
	do
		for arch in $ARCHLIST
		do
			if [ -f "sf2gh.$os.$arch" ]
			then
				echo "removing $os.$arch"
				rm "sf2gh.$os.$arch"
			fi
		done
	done
}

REMOVE=0
OSLIST="linux darwin"
ARCHLIST="amd64 386 arm"

while getopts "ra:o:" opt
do
	case "$opt" in
		"r")
			REMOVE=1
			;;
		"a")
			ARCHLIST="$OPTARG"
			;;
		"o")
			OSLIST="$OPTARG"
			;;
		[?])
			exit 1
			;;
	esac
done

if [ $REMOVE -eq 1 ];then
	remove
else
	build "$ARCHLIST" "$OSLIST"
fi
