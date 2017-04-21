# Transit Route Planner

# Getting Started

These instructions are for setting up a local development environment.

## Prerequisites

You'll need to [install Go](https://golang.org/doc/install) and set up your Go workspace. [Instructions here](https://golang.org/doc/code.html).

~~~
wget https://storage.googleapis.com/golang/go1.8.1.linux-amd64.tar.gz
tar -C /usr/local -xvf go1.8.1.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
mkdir $HOME/go
cd $HOME/go
~~~

This project uses [protocol buffers](https://developers.google.com/protocol-buffers/docs/overview). Install from repo [here](https://github.com/golang/protobuf).

## Installing

~~~
cd $HOME/go
git clone git@github.com:xianny/mta-gtfs.git
cd mta-gtfs
go build
~~~

Run with `./mta-gtfs`
