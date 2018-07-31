CFLAGS=-g -fPIC -m64 -Wall -Wextra -Wno-unused-parameter -Wno-unused-result -O2

CXXFLAGS=-g -std=c++11 -fno-rtti -fPIC -m64 -Wall -Wextra -Wno-unused-parameter -Wno-unused-result -O2
CXXFLAGS+=-I./include

LDFLAGS=-fvisibility=hidden

LIB_PATH=-L./libv8
LIBS=-lv8_libbase -lv8_libplatform -licui18n -licuuc

CXX=g++
LD=g++

vm: vm.cc.o
	$(LD) -g -shared $(LDFLAGS) $^ -o libvm.so $(LIB_PATH) $(LIBS)

%.cc.o: %.cc
	$(CXX) $(CXXFLAGS) -c $< -o $<.o

install:
	-mv libvm.so ./libv8

clean:
	-rm -rf *.o
	-rm -rf *.so