file=stdout
async=false

run:
	go build -o lig && sudo ./lig -out=$(file) -async=$(async)