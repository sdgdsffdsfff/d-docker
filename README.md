dea-docker
====
docker 后端管理系统

==============
install
==============

1. 下载go语言包
	download go1.2.linux-adm64.tar.gz
	
2. 配置go环境变量
	export GIT_SSL_NO_VERIFY=1
	export GOROOT=/export/service/go
	export GOPATH=/export/service/gopath
	export GOBIN=$GOROOT/bin
	export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
	
3.提前下载依赖包到 $GOPATH/src
	git clone http://icode.jd.com/cdlxyong/go-dockerclient.git 到 $GOPATH/src/icode.jd.com/cdlxyong
	下载 gosteno 到 $GOPATH/src/github.com/cloudfoundry/
	下载 yagnats 到 $GOPATH/src/github.com/cloudfoundry/
	
4.下载 dea-docker
	cd $GOPATH/src && git clone http://icode.jd.com/cdlxyong/dea-docker.git

5:编译
	cd $GOPATH/src/dea-docker/src && go get -v ./...
	
6:修改配置文件
	默认配置在 $GOPATH/src/dea-docker/src/example_config/config.yml
	
7:启动
	dea -c config_path
	or 
	nohup dea -c config.yml > /export/home/jae/dea-docker.out

	

