#!/bin/bash

cd /chemotion/app

ifEmailExists=$(echo "User.find_by(email:'$1')" | bundle exec rails c | tail -1)

if [ "$ifEmailExists" == "nil" ]; then
    echo "No record found"
else
    userUpdated=$(echo "User.find_by(email:'$1').update(password:'${2:-chemotion}', first_name:'$3', last_name:'$4', type:'Person', name_abbreviation:'$5')" | bundle exec rails c | tail -1)
    if [[ "$userUpdated" == "true" ]]; then
        echo "User updated successfuly"
    else
        echo "Please use unique abbreviation"
    fi
fi