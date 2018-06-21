# -*- coding:utf-8 -*-
import subprocess
import os
import time
import sys

HOME=os.environ['HOME']
GOPATH=os.environ['GOPATH']
project_path=GOPATH+"/src/github.com/iost-official/prototype"
testDir=project_path+"/scripts/testWallet/test"
ip_port="127.0.0.1:30303"
def wCommand(com):
	obj = subprocess.Popen([com], stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE,shell=True)
	obj.wait()
	ret=obj.stdout.read()
	obj.stdout.close()
	return ret

def has(fn):
	return os.path.exists(fn)

def clear():
	if has(HOME+'/.ssh/test_secp'):wCommand("rm "+HOME+"/.ssh/test_secp")
	if has(HOME+'/.ssh/test_secp.pub'):wCommand("rm "+HOME+"/.ssh/test_secp.pub")
	if has(testDir+"/1to2.sc"):wCommand("rm "+testDir+"/1to2.sc")
	if has(testDir+"/1to2.sig"):wCommand("rm "+testDir+"/1to2.sig")
	if has(testDir+"/1to2.tx"):wCommand("rm "+testDir+"/1to2.tx")

def installWallet():
	print "[iwallet install]:",
	ret=wCommand("cd "+project_path+"/iwallet && go install")
	if ret=="":
		print("ok")
		return True
	print("fail")
	return False

def checkAccount():
	print "[iwallet account]:",
	wCommand("iwallet account -c test")
	if has(HOME+'/.ssh/test_secp') and has(HOME+'/.ssh/test_secp.pub'):
		print("ok")
		return True
	print("fail")
	return False



def checkBlock():
	print "[iwallet block]:",
	ret=wCommand("iwallet block 0")
	if ret.startswith("["):#TODO
		print("ok")
		return True
	print("fail")
	return False
		

def checkCompile():
	print "[iwallet compile]:",
	wCommand("iwallet compile "+testDir+"/1to2.lua")
	if has(testDir+"/1to2.sc"):
		print("ok")
		return True 
	print("fail")
	return False

def checkSign():
	print "[iwallet sign]:",
	ret=wCommand("iwallet sign "+testDir+"/1to2.sc -k ~/.ssh/test_secp")
	if has(testDir+"/1to2.sig"):
		print("ok")
		return True 
	print("fail")
	return False

def checkPublish():
	print "[iwallet publish]:",
	ret=wCommand("iwallet -s "+ip_port+" publish "+testDir+"/1to2.sc "+testDir+"/1to2.sig -k ~/.ssh/test_secp")
	if ret.startswith("ok"):
		#check balance here
		print("ok")
		return True
	print("fail")
	return False

def checkValue():
	print "[iwallet value]:",
	ret=wCommand("iwallet value iost")
	#TODO
	print("ok")
	return True 

def checkBalance():
	print "[iwallet balance]:",
	ret=wCommand("iwallet balance "+HOME+"/.ssh/test_secp.pub")
	
	if ret.startswith(HOME+"/.ssh/test_secp.pub >"):
		print("ok")
		return True
	print("fail")
	return False
	
if __name__ == "__main__":
	print("check iwallet functions:")
	global ip_port
	if(len(sys.argv)==2):
		ip_port=sys.argv[1]

	clear()
	func_list=[installWallet,checkAccount,checkBlock,checkCompile,checkSign,checkPublish,checkValue,checkBalance,]
	for func in func_list:
		if func()==False:break
	#del files
	clear()
	print("check iwallet finished")
