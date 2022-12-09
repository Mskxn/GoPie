import os

path_list = [
    "./testdata/project/blocking/cockroach/10214/cockroach10214_test.go",
    "./testdata/project/blocking/cockroach/1055/cockroach1055_test.go",
    "./testdata/project/blocking/cockroach/10790/cockroach10790_test.go",
    "./testdata/project/blocking/cockroach/13197/cockroach13197_test.go",
    "./testdata/project/blocking/cockroach/13755/cockroach13755_test.go",
    "./testdata/project/blocking/cockroach/1462/cockroach1462_test.go",
    "./testdata/project/blocking/cockroach/16167/cockroach16167_test.go",
    "./testdata/project/blocking/cockroach/18101/cockroach18101_test.go",
    "./testdata/project/blocking/cockroach/2448/cockroach2448_test.go",
    "./testdata/project/blocking/cockroach/24808/cockroach24808_test.go",
    "./testdata/project/blocking/cockroach/25456/cockroach25456_test.go",
    "./testdata/project/blocking/cockroach/35073/cockroach35073_test.go",
    "./testdata/project/blocking/cockroach/35931/cockroach35931_test.go",
    "./testdata/project/blocking/cockroach/3710/cockroach3710_test.go",
    "./testdata/project/blocking/cockroach/584/cockroach584_test.go",
    "./testdata/project/blocking/cockroach/6181/cockroach6181_test.go",
    "./testdata/project/blocking/cockroach/7504/cockroach7504_test.go",
    "./testdata/project/blocking/cockroach/9935/cockroach9935_test.go",
    "./testdata/project/blocking/etcd/10492/etcd10492_test.go",
    "./testdata/project/blocking/etcd/5509/etcd5509_test.go",
    "./testdata/project/blocking/etcd/6708/etcd6708_test.go",
    "./testdata/project/blocking/etcd/6857/etcd6857_test.go",
    "./testdata/project/blocking/etcd/6873/etcd6873_test.go",
    "./testdata/project/blocking/etcd/7443/etcd7443_test.go",
    "./testdata/project/blocking/etcd/7492/etcd7492_test.go",
    "./testdata/project/blocking/etcd/7902/etcd7902_test.go",
    "./testdata/project/blocking/grpc/1275/grpc1275_test.go",
    "./testdata/project/blocking/grpc/1353/grpc1353_test.go",
    "./testdata/project/blocking/grpc/1424/grpc1424_test.go",
    "./testdata/project/blocking/grpc/1460/grpc1460_test.go",
    "./testdata/project/blocking/grpc/3017/grpc3017_test.go",
    "./testdata/project/blocking/grpc/660/grpc660_test.go",
    "./testdata/project/blocking/grpc/795/grpc795_test.go",
    "./testdata/project/blocking/grpc/862/grpc862_test.go",
    "./testdata/project/blocking/hugo/3251/hugo3251_test.go",
    "./testdata/project/blocking/hugo/5379/hugo5379_test.go",
    "./testdata/project/blocking/istio/16224/istio16224_test.go",
    "./testdata/project/blocking/istio/17860/istio17860_test.go",
    "./testdata/project/blocking/istio/18454/istio18454_test.go",
    "./testdata/project/blocking/kubernetes/10182/kubernetes10182_test.go",
    "./testdata/project/blocking/kubernetes/11298/kubernetes11298_test.go",
    "./testdata/project/blocking/kubernetes/13135/kubernetes13135_test.go",
    "./testdata/project/blocking/kubernetes/1321/kubernetes1321_test.go",
    "./testdata/project/blocking/kubernetes/25331/kubernetes25331_test.go",
    "./testdata/project/blocking/kubernetes/26980/kubernetes26980_test.go",
    "./testdata/project/blocking/kubernetes/30872/kubernetes30872_test.go",
    "./testdata/project/blocking/kubernetes/38669/kubernetes38669_test.go",
    "./testdata/project/blocking/kubernetes/5316/kubernetes5316_test.go",
    "./testdata/project/blocking/kubernetes/58107/kubernetes58107_test.go",
    "./testdata/project/blocking/kubernetes/62464/kubernetes62464_test.go",
    "./testdata/project/blocking/kubernetes/6632/kubernetes6632_test.go",
    "./testdata/project/blocking/kubernetes/70277/kubernetes70277_test.go",
    "./testdata/project/blocking/moby/17176/moby17176_test.go",
    "./testdata/project/blocking/moby/21233/moby21233_test.go",
    "./testdata/project/blocking/moby/25384/moby25384_test.go",
    "./testdata/project/blocking/moby/27782/moby27782_test.go",
    "./testdata/project/blocking/moby/28462/moby28462_test.go",
    "./testdata/project/blocking/moby/29733/moby29733_test.go",
    "./testdata/project/blocking/moby/30408/moby30408_test.go",
    "./testdata/project/blocking/moby/33293/moby33293_test.go",
    "./testdata/project/blocking/moby/33781/moby33781_test.go",
    "./testdata/project/blocking/moby/36114/moby36114_test.go",
    "./testdata/project/blocking/moby/4395/moby4395_test.go",
    "./testdata/project/blocking/moby/4951/moby4951_test.go",
    "./testdata/project/blocking/moby/7559/moby7559_test.go",
    "./testdata/project/blocking/serving/2137/serving2137_test.go",
    "./testdata/project/blocking/syncthing/4829/syncthing4829_test.go",
    "./testdata/project/blocking/syncthing/5795/syncthing5795_test.go",
]
file_names = map(lambda x: x.split("/")[-1], path_list)
fuzz_names = map(lambda x: "FuzzGen" + x.captialize(), map(lambda x: x.split("_test.go")[0], file_names))

def inst_file(fn):
    print(f"handle {fn}", end="")
    os.system(f"./sw.exe --file {fn}")
    print("\tok")

def fuzz_target(fname, fn):
    print(f"test {fn}", end="")
    os.system(f"go test -fuzz {fname}")
    print("ok")

for fn in path_list:
    inst_file(fn)
