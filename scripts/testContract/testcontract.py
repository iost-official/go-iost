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
def wCommand(com):
	obj = subprocess.Popen([com], stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE,shell=True)
	obj.wait()
	ret=obj.stdout.read()
	obj.stdout.close()
	return ret

def has(fn):
	return os.path.exists(fn)

def Buildwallet():
 	ret=wCommand("cd "+project_path+"iwallet;go build")
	ret=wCommand("rm -f "+cur_path+"iwallet")
	ret=wCommand("cp "+project_path+"iwallet/iwallet "+cur_path)
	return True

def Compile(fileName, account):
	wCommand("rm -f "+cur_path+"test/"+fileName+".sc")
        wCommand(cur_path+"iwallet -s "+server_addr+" compile -n "+str(random.randint(0,sys.maxint))+" "+cur_path+"test/"+fileName+".lua")
	if has(cur_path+"test/"+fileName+".sc"):
		#print("ok")
		return True 
	#print("fail")
	return False

def Sign(fileName, account):
	#print "[iwallet sign]:",
	wCommand("rm -f "+cur_path+"test/"+fileName+".sig")
        print(cur_path+"iwallet -s "+ server_addr + " sign "+cur_path+"test/"+fileName+".sc -k ~/.ssh/"+account+"_secp")
        ret=wCommand(cur_path+"iwallet -s "+ server_addr + " sign "+cur_path+"test/"+fileName+".sc -k ~/.ssh/"+account+"_secp")
	if has(cur_path+"test/"+fileName+".sig"):
		#print("ok")
		return True 
	#print("fail")
	return False

def Publish(fileName, account):
	print "[iwallet publish]:",
        print(cur_path+"iwallet -s "+server_addr+" publish "+cur_path+"test/"+fileName+".sc -k ~/.ssh/"+account+"_secp")
        ret=wCommand(cur_path+"iwallet -s "+server_addr+" publish "+cur_path+"test/"+fileName+".sc -k ~/.ssh/"+account+"_secp")
        print ret
	if ret.startswith("ok"):
		#check balance here
		#print("ok")
		return True
	#print("fail")
	return False

def PublishTx(fileName, account):
	flist=[Compile,Publish,]
	for func in flist:
		print func
		if func(fileName, account)==False:
			return "FAIL"
	return "SUCCESS"

genesisPubkey="2BibFrAhc57FAd3sDJFbPqjwskBJb5zPDtecPWVRJ1jxT"
genesisSeckey="BRpwCKmVJiTTrPFi6igcSgvuzSiySd7Exxj7LGfqieW9"

if __name__ == "__main__":
	if has(HOME+"/.ssh")==False:
		wCommand("mkdir ~/.ssh")
	f1=open(HOME+"/.ssh/test_secp","w")
	f1.write(genesisSeckey)
	f1.close()
	
	f2=open(HOME+"/.ssh/test_secp.pub","w")
	f2.write(genesisPubkey)
	f2.close()

	com=sys.argv[1]
	if com=="publishtx":
                fileName = sys.argv[2]
                account = sys.argv[3]
		print(PublishTx(fileName, account))
		sys.exit(0)

