#!/bin/bash

echo "Installing tayniserver"
go install -v taynireporter.go

FILE=/usr/local/bin/taynireporter

if [ -f $FILE ]; then
    echo "File $FILE exists."
else
    echo "File $FILE does not exist."
    sudo ln -s $GOPATH/bin/taynireporter /usr/local/bin/taynireporter 
    fi

echo "Installing service..."
sudo cp $GOPATH/src/github.com/lagarciag/tayni/taynireporter/taynireporter.service /etc/systemd/system/
echo "Enabling service..."
sudo systemctl enable taynireporter.service

