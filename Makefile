frontend-build:
	cd frontend && npm install && ng build --configuration production

build: frontend-build
	go build -o monitor main.go

test:
	cd frontend && npm run test 

clean:
	rm -f monitor
	rm -rf frontend/dist

install:
	cp monitor /usr/local/bin/
	sudo setcap cap_net_raw+eip /usr/local/bin/monitor