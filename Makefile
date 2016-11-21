FILES = main.go dropbox.go rtorrent.go

run:
	go run $(FILES)

fmt:
	go fmt $(FILES)

build:
	go build -o bin/slackbot-x64 $(FILES)
	GOARCH=386 go build -o bin/slackbot-x86 $(FILES)

deploy:
	rsync -av --progress bin/slackbot-x64 mjolnir:~/slackbot/slackbot
	ssh mjolnir -t "sudo supervisorctl restart slackbot"
	rsync -av --progress bin/slackbot-x86 bifrost:~/slackbot/slackbot
	ssh bifrost -t "sudo supervisorctl restart slackbot"

build-and-deploy: build deploy
