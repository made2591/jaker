#!/bin/sh

export ACTUAL_BYTES_LIMIT_SIZE=1073741824
#export ACTUAL_BYTES_LIMIT_SIZE=2147483648

#https://github.com/vjeantet/alerter

bytesToHuman() {
  b=${1:-0}; d=''; s=0; S=(Bytes {K,M,G,T,E,P,Y,Z}iB)
  while ((b > 1024)); do
    d="$(printf ".%02d" $((b % 1024 * 100 / 1024)))"
    b=$((b / 1024))
    let s++
  done
  echo "$b$d ${S[$s]}"
}

jakerAlerter() {
  ./alerter -title Jaker -subtitle $1 -message $2 -closeLabel No -actions Yes -appIcon $3
}

jakerDismiss() {
  ./alerter -message "Dismis this notification?" -actions "Now","Later today","Tomorrow" -dropdownLabel "When?" -appIcon $1
}

jakerModifier() {
  ./alerter -reply -message $1 -title $2 -appIcon $3
}

while true
do

  ACTUAL_BYTES_SIZE=$(./dockerLRSize size)
  ACTUAL_HUREA_SIZE=$(bytesToHuman $(echo $ACTUAL_BYTES_SIZE))
  ACTUAL_HUREA_LIMIT_SIZE=$(bytesToHuman $(echo $ACTUAL_BYTES_LIMIT_SIZE))

  #echo $ACTUAL_BYTES_SIZE
  #echo $ACTUAL_HUREA_SIZE
  #echo $ACTUAL_BYTES_LIMIT_SIZE
  #echo $ACTUAL_HUREA_LIMIT_SIZE

  if [ $ACTUAL_BYTES_SIZE \> $ACTUAL_BYTES_LIMIT_SIZE ];
  then 
    #echo "$ACTUAL_BYTES_SIZE is greater than $ACTUAL_BYTES_LIMIT_SIZE";
    #response=$(jakerAlerter "Limit overcome!" "Remove unused images?" "./img/docker.png")
    response=$(jakerDismiss "./img/docker.png")
    if [ "$response" == "Yes" ]; then
      docker image prune --all > /dev/null 2>&1
      ACTUAL_BYTES_SIZE=$(./dockerLRSize size)
      if [ $ACTUAL_BYTES_SIZE \> $ACTUAL_BYTES_LIMIT_SIZE ];
      then 
        response=$(jakerModifier "Cleaning not sufficient" "Do you want to change the limit?" "./img/docker.png")        
        echo $response
      fi;
    fi;
  fi;
  sleep 1
done