#! /usr/bin/env python
# -*- coding:utf-8 -*-
import subprocess
import os
import time
import sys
import random
from flask import Flask,request

HOME=os.environ['HOME']
GOPATH=os.environ['GOPATH']
cur_path=GOPATH+"/src/github.com/iost-official/prototype/scripts/depatch/"
project_path=GOPATH+"/src/github.com/iost-official/prototype/"
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
	return (True,"build iwallet success")

genesisPubkey="2BibFrAhc57FAd3sDJFbPqjwskBJb5zPDtecPWVRJ1jxT"
genesisSeckey="BRpwCKmVJiTTrPFi6igcSgvuzSiySd7Exxj7LGfqieW9"

def construct(_from,_to,money):
	f=open(cur_path+"test/1to2.lua","w")
	f.writelines([
		'--- main 合约主入口\n',
		'-- server1转账server2\n',
		'-- @gas_limit 10000\n',
		'-- @gas_price 0.001\n',
		'-- @param_cnt 0\n',
		'-- @return_cnt 0\n',
		'function main()\n',
		'	Transfer("'+_from+'","'+_to+'",'+str(money)+')\n',
		'end--f\n',
	])
	f.close()

def Compile():
	#print "[iwallet compile]:",
	wCommand("rm -f "+cur_path+"test/1to2.sc")
	ret=wCommand(cur_path+"iwallet compile -n "+str(random.randint(0,sys.maxint))+" "+cur_path+"test/1to2.lua")
	if has(cur_path+"test/1to2.sc"):
		return (True,ret)
	return (False,ret)

def Sign():
	#print "[iwallet sign]:",
	wCommand("rm -f "+cur_path+"test/1to2.sig")
	ret=wCommand(cur_path+"iwallet sign "+cur_path+"test/1to2.sc -k ~/.ssh/genesis_secp")
	if has(cur_path+"test/1to2.sig"):
		#print("ok")
		return (True,ret)
	#print("fail")
	return (False,ret)

def Publish():
	#print "[iwallet publish]:",
	ret=wCommand(cur_path+"iwallet publish -s 52.56.118.10:30303 "+cur_path+"test/1to2.sc "+cur_path+"test/1to2.sig -k ~/.ssh/genesis_secp")
	if ret.startswith("ok"):
		#check balance here
		return (True,ret[3:-1])
	return (False,ret)

#send money for someone,used in blockchain explorer
def sendonetx(_to,money):
	construct(genesisPubkey,_to,money)
	f1=open(HOME+"/.ssh/genesis_secp","w")
	f1.write(genesisSeckey)
	f1.close()
	
	f2=open(HOME+"/.ssh/genesis_secp.pub","w")
	f2.write(genesisPubkey)
	f2.close()

	if has(cur_path+"iwallet")==False:Buildwallet()
	flist=[Compile,Sign,Publish,]
	ret=(True,"")
	for func in flist:
		ret=func()
		if ret[0]==False:
			print(func)
			return ret
	return ret

app = Flask(__name__)
@app.route('/')
def indexpage():
	return "hello iost"
@app.route('/givemoney',methods=['GET',])
def givemoney():
	print("givemoney begin")
	if request.method=='GET':
		args=request.args.to_dict()
		user=args['user']
		money=args['money']
		ret=sendonetx(user,float(money))
		return ret[1]


if __name__ == '__main__':
    app.run(host = '0.0.0.0',port = 8080,)
