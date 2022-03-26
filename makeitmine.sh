#!/bin/bash
# http://redsymbol.net/articles/unofficial-bash-strict-mode/
set -euo pipefail
IFS=$'\n\t'

cwd=${PWD}
basepath=$(dirname -- ${BASH_SOURCE[0]})

thisfile=`basename "$0"`
fontcolorgreen=$(tput setaf 2)
fontbold=$(tput bold)
fontnormal=$(tput sgr0)

existingmodulename="github.com/bnkamalesh/goapp"
modulename=''
a_flag=false

print_usage() {
  printf """
Usage:
./makeitmine.sh -n <module name> -a -e example.org/me/myapp

Options
=======
-n    Parameter accepts the new module name, and is mandatory
-a    Optional flag if set, will do the following:
        - remove .git directory
        - empty README.md
        - remove .travis.yml
-e    Optional parameter accepts an existing module name to be replaced. Default is 'github.com/bnkamalesh/pusaki'

  """
}

# https://stackoverflow.com/a/7069755
while getopts 'an:e:' flag; do
  case "${flag}" in
    a) a_flag=true;;
    n) modulename="${OPTARG}" ;;
    e) existingmodulename="${OPTARG}" ;;
    *) print_usage
       echo "exiting?"
       exit 1 ;;
  esac
done

if [ -z "$modulename" ]
then
  print_usage
  exit 2;
fi

if [ -z "$existingmodulename" ]
then
  print_usage
  exit 3;
fi

totalsteps="2"
if [ $a_flag = true ]
then
  totalsteps="3"
fi

currentStep=1

printf "${fontbold}[$currentStep/$totalsteps] Setting up your module '$modulename'${fontnormal}\n"
# Replacing . with \. to escape . while being used in regex
unescapedmodule=$modulename
modulename=${modulename//./\\.}
unescapedexistingmodule=$existingmodulename
existingmodulename=${existingmodulename//./\\.}

# using ; as separator in sed, because `/` is prevalent in most Go module paths
replacer="s;$existingmodulename;$modulename;g"
echo " - Replacing module '$unescapedexistingmodule' with '$unescapedmodule'"
grep -rl $existingmodulename ${basepath}/ --exclude-dir=.git --exclude-dir=vendor --exclude=README.md --exclude=$thisfile | xargs sed -i $replacer
printf "${fontcolorgreen}${fontbold}= Done${fontnormal}\n"

if [ $a_flag = true ]
then
  ((currentStep++))
  printf "\n${fontbold}7 Deleting .git and emptying README${fontnormal}\n"
  rm -rf ${basepath}/.git ${basepath}/.travis.yml
  echo "" > ${basepath}/README.md
  printf "${fontcolorgreen}${fontbold}= Done${fontnormal}\n"
fi

((currentStep++))
printf "\n${fontbold}[$currentStep/$totalsteps] Go clean-up ${fontnormal}\n"
cd $basepath
go mod tidy
go mod verify
printf "${fontcolorgreen}${fontbold}= Done${fontnormal}\n"

printf "\n${fontcolorgreen}${fontbold}=== Done [$currentStep/$totalsteps]${fontnormal}\n\n"
exit 0;
