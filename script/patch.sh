GOROOT=$(go env GOROOT)
SRCPATH=$GOROOT/src

cp ./patch/runtime/chan.go.patch $SRCPATH/runtime/chan.go
cp ./patch/sync/mutex.go.patch $SRCPATH/sync/mutex.go
cp ./patch/runtime/runtime2.go.patch $SRCPATH/runtime/runtime2.go

