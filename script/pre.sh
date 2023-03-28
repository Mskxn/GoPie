cd /tool

rm -r ./testbins
cd $1
git checkout -- *
cd ..

go build -o ./bin ./cmd/...
./bin/fuzz --task inst --path $1
./script/patch.sh
./bin/fuzz --task bins --path $1