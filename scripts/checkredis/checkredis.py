#! /usr/bin/env python
# -*- coding:utf-8 -*-

import subprocess
import os
import time
import json

server_addr = [["Tokyo", "18.179.143.193","1"],
["London", "52.56.118.10", "2"],
["Singapore", "13.228.206.188", "3"],
["Mumbai", "13.232.96.221", "4"],
["Frankfort", "18.184.239.232", "5"],
["Seoul", "13.124.172.86", "6"],
["Montreal", "52.60.163.60","7"]]

#fout = open("out.txt","w")

balance_map = {}
bn_map = {}
while True:
    min_bn = -1
    for item in server_addr:
        obj = subprocess.Popen("curl -s --connect-timeout 3 -XPOST {}:{}/scripts -d \"cmd=checkredis.sh\"".format(item[1], 30310), 
            shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        rtn, _ =obj.communicate()
        lines = rtn.split("\n")[:-1]
        if len(lines)%2 ==1:
            lines[1] += "\n"+lines[2]
            lines[2:-1] = lines[3:]
            lines = lines[:-1]
        if len(lines) < 2:
            print("Error: %s ", item[0])
            continue

        bn = int(lines[0])
        print item[0], bn, len(lines)/2 -1

        if not bn in bn_map:
            bn_map[bn] = {}
        if item[2] in bn_map[bn]:
            continue
        bn_map[bn][item[2]] = 1

        if min_bn == -1 or bn<min_bn:
            min_bn = bn

        bh = lines[1]
        if not bn in balance_map:
            balance_map[bn] = {}
        if not bh in balance_map[bn]:
            balance_map[bn][bh] = {}

        for i in range(2, len(lines), 2):
            key, val = lines[i:i+2]
            val = float(val)
            #print key, val
            if not key in balance_map[bn][bh]:
                balance_map[bn][bh][key] = {}
            if not val in balance_map[bn][bh][key]:
                balance_map[bn][bh][key][val]= [] 
            balance_map[bn][bh][key][val].append(item[2])
    
    boo = True
    del_map = []
    for bn in balance_map:
        if bn < min_bn:
            del_map.append(bn)
            continue
        #print balance_map[bn]
        if len(balance_map[bn]) >1:
            boo = False
            print("Error: %s fork", bn)
        for bh in balance_map[bn]:
            for key in balance_map[bn][bh]:
                if len(balance_map[bn][bh][key])>1:
                    boo = False
                    print key
                    print balance_map[bn][bh][key]
    for bn in del_map:
        bn_map.pop(bn)
        balance_map.pop(bn)
    print(len(balance_map))
    if boo:
        print "OK!"
    print 

    time.sleep(2)
