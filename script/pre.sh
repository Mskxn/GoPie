cd /tool

rm -r ./testbins
cd $1
git checkout -- *
go get go.uber.org/goleak
cd ..

go build -o ./bin ./...
./bin/fuzz --task inst --path $1
./script/patch.sh
./bin/fuzz --task bins --path $1