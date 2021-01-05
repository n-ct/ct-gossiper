# ct-gossiper

This version is for testing in Deterlab.
The gossip server uses GLOG so check here -> https://godoc.org/github.com/golang/glog for the GLOG options.

EXAMPLE OF HOW TO USE -

  > server.exe -log_dir=./log -port 2000 -n 5 -ip 2

  This will start the server in port 2000, there is a total of 5 nodes in the experiment, and the ip of this server is 10.1.10.2

FLAGS
-port: port number where the server will be running. 1000 if not selected. This assumes that all the other gossipers will be running the same port.
-ip: last values in the ip address of the server. During the experiments, the ip addresses have the format 10.1.10.xxx; the ip 10.1.10.1 is reserved for the router. This flag is mandatory.
-n: total number of gossipers in the experiment. This flag is mandatory and must be bigger than 1

Refer to https://github.com/n-ct/ct-gossiper for original version.
