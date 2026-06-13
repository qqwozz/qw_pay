.PHONY: build run clean

fraud_engine: fraud_engine.cpp
	g++ -std=c++17 -O2 -o fraud_engine fraud_engine.cpp -lhiredis -lpthread

run: fraud_engine
	./fraud_engine

clean:
	rm -f fraud_engine
