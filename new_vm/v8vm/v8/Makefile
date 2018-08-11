CFLAGS=-g -fPIC -m64 -Wall -Wextra -Wno-unused-parameter -Wno-unused-result -O2

CXXFLAGS=-g -std=c++11 -fno-rtti -fPIC -m64 -Wall -Wextra -Wno-unused-parameter -Wno-unused-result -O2

LDFLAGS=-fvisibility=hidden

LIBS=-lv8_libbase -lv8_libplatform

OS_NAME = $(shell uname -s)
LC_OS_NAME = $(shell echo $(OS_NAME) | tr '[A-Z]' '[a-z]')

ifeq ($(LC_OS_NAME), darwin)
	CC=clang
	CXX=clang++
	LD=clang++
	LIB_SUFFIX=.dylib
	CXXFLAGS+=-I./include/_darwin_amd64
	LIBS+=-lv8_base -lv8_init -lv8_initializers -lv8_libsampler -lv8_snapshot
	LIB_PATH=./libv8/_darwin_amd64
	INSTALL=install
	INSTALL_PATH=./libv8/_darwin_amd64
	LDCONFIG=
else
	CC=gcc
	CXX=g++
	LD=g++
	LIB_SUFFIX=.so
	CXXFLAGS+=-I./include/_linux_amd64
	LIBS+=-licui18n -licuuc
	LIB_PATH=./libv8/_linux_amd64
	INSTALL=sudo install
	INSTALL_PATH=./libv8/_linux_amd64
	LDCONFIG=sudo /sbin/ldconfig
endif

vm: vm.cc.o require.cc.o storage.cc.o blockchain.cc.o sandbox.cc.o instruction.cc.o
	$(LD) -g -shared $(LDFLAGS) $^ -o libvm$(LIB_SUFFIX) -L$(LIB_PATH) $(LIBS)

%.cc.o: %.cc
	$(CXX) $(CXXFLAGS) -c $< -o $<.o

install:
	-mv libvm$(LIB_SUFFIX) $(INSTALL_PATH)
	$(INSTALL) $(INSTALL_PATH)/*$(LIB_SUFFIX) /usr/local/lib
	$(LDCONFIG)

clean:
	-rm -rf *.o
	-rm -rf *$(LIB_SUFFIX)