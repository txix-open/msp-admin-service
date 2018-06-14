#!/usr/bin/env bash

START=`date +%N`

#########################
# The command line help #
#########################
display_help() {
    echo "Usage: $0 [option...]" >&2
    echo
    echo "   -ma, --major           increase major version, X.0.0. Minor and build versions will set eq 0"
    echo "   -mi, --minor           increase minor version, 0.X.0. Build version will set eq 0"
    echo
    echo "Run without option means that will increase build version 0.0.X"
    echo
    # echo some stuff here for the -a or --add-options
    exit 1
}

INCREASE_MAJOR_VERSION=0
INCREASE_MINOR_VERSION=0

################################
# Check if parameters options  #
# are given on the commandline #
################################
case "$1" in
  -ma | --major)
    INCREASE_MAJOR_VERSION=1
    ;;
  -mi | --minor)
    INCREASE_MINOR_VERSION=1
    ;;
  -h | --help)
    display_help  # Call your function
    exit 0
    ;;
  --) # End of all options
    exit 1
    ;;
  -*)
    echo "Error: Unknown option: $1" >&2
    exit 1
    ;;
  *)  # No more options
    ;;
esac

VERSION="0.1.0"
PREVIOUS_VERSION="0.1.0"
MAJOR_VERSION=0
MINOR_VERSION=1
BUILD_VERSION=0

FILE=".version"
FILE_PATH="./$FILE"
if [ ! -f "$FILE_PATH" ]
then
    `echo ${VERSION} > ${FILE_PATH}`
else
    PREVIOUS_FILE_VERSION=`head -n 1 ${FILE_PATH}`
    MAJOR_FILE_VERSION="$( cut -d'.' -f1 <<<${PREVIOUS_FILE_VERSION} )"
    MINOR_FILE_VERSION="$( cut -d'.' -f2 <<<${PREVIOUS_FILE_VERSION} )"
    BUILD_FILE_VERSION="$( cut -d'.' -f3 <<<${PREVIOUS_FILE_VERSION} )"
    if [ -z "$MAJOR_FILE_VERSION" ] || [ -z "$MINOR_FILE_VERSION" ] || [ -z "$BUILD_FILE_VERSION" ]
    then
        echo
        printf 'Wrong content in the file .version: %s, must contain 3 segments with point as a separator' ${PREVIOUS_FILE_VERSION}
        echo
    else
        MAJOR_VERSION=${MAJOR_FILE_VERSION}
        MINOR_VERSION=${MINOR_FILE_VERSION}
        BUILD_VERSION=${BUILD_FILE_VERSION}
        PREVIOUS_VERSION=${PREVIOUS_FILE_VERSION}
        if [ ${INCREASE_MAJOR_VERSION} -ne 0 ]
        then
            MAJOR_VERSION=$((MAJOR_VERSION + 1))
            MINOR_VERSION=0
            BUILD_VERSION=0
        elif [ ${INCREASE_MINOR_VERSION} -ne 0 ]
        then
            MINOR_VERSION=$((MINOR_VERSION + 1))
            BUILD_VERSION=0
        else
            BUILD_VERSION=$((BUILD_VERSION + 1))
        fi
    fi
fi

VERSION="$MAJOR_VERSION.$MINOR_VERSION.$BUILD_VERSION"
`echo ${VERSION} > ${FILE_PATH}`

divider=====================
divider=$divider$divider$divider$divider$divider
width=100
format="| ${blue}%-20s ${normal} | %-72s |\n"

export GOPATH="$PWD/../../"
NOW="+%Y-%m-%d %H:%M:%S"
NOW=$(date "$NOW")
echo ${NOW}
CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -v -ldflags="-X 'main.version=$VERSION' -X 'main.date=$NOW'"

END=`date +%N`
DURATION=`echo "scale=2; $(( ($END-$START) ))/1000000" |bc`
echo
printf "%$width.${width}s\n" "$divider"
printf "$format" "VERSION:" "$PREVIOUS_VERSION -> $VERSION"
printf "$format" "DATE:" "$NOW"
printf "$format" "EXECUTED TIME:" "$DURATION ms"
printf "%$width.${width}s\n" "$divider"
echo
