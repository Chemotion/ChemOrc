#!/bin/bash

cd /chemotion/app

ifEmailExists=$(echo "User.find_by(email:'$1')" | bundle exec rails c | tail -1)

if [ $ifEmailExists == "nil" ]; then
    echo "No record found";
else
    echo "User.find_by(email:'$1').destroy" | bundle exec rails c | tail -1 && echo "User deleted succesfuly."; fi