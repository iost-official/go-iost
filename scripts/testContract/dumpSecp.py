#! /usr/bin/env python
# -*- coding:utf-8 -*-
import subprocess
import os
import time
import sys
import random 

HOME='/home/wangyu'
GOPATH='/home/wangyu/gocode'
cur_path=GOPATH+"/src/github.com/iost-official/prototype/scripts/testContract/"
project_path=GOPATH+"/src/github.com/iost-official/prototype/"
server_addr='127.0.0.1:30313'

genesisPubkey="2BibFrAhc57FAd3sDJFbPqjwskBJb5zPDtecPWVRJ1jxT"
genesisSeckey="BRpwCKmVJiTTrPFi6igcSgvuzSiySd7Exxj7LGfqieW9"

def has(fn):
	return os.path.exists(fn)

def dump(name, pubkey, seckey):
	f1=open(HOME+"/.ssh/"+name+"_secp.pub","w")
	f1.write(pubkey)
	f1.close()
	
	f2=open(HOME+"/.ssh/"+name+"_secp","w")
	f2.write(seckey)
	f2.close()

	if has(HOME+"/.ssh/"+name+"_secp") and has(HOME+"/.ssh/"+name+"_secp.pub"):
		return True 
	return False

if __name__ == "__main__":
	if has(HOME+"/.ssh")==False:
		wCommand("mkdir ~/.ssh")

        print(dump("a", "gvCQNmkuA6AwdddRMSUg6jr8W7swKWAnhEY3cAthj9bX", "8a11FHFjvWDbtx4gR4JJWBwXDVwTwDFDMv7F8J6wFhyN"))
        print(dump("b", "2538yUDuKTLaXqCTFS1tfVmMEL4dVnzLDWChoMdoxgCa4", "YJXWyJkMiSAYeHkUJukCUPW8srfXmCFF148isFz2RdY"))
