# spark-sql-on-k8s
最简单的 spark sql on kubernetes 生产环境部署方案

# 架构和设计
详细架构和设计内容，请参考 https://zhuanlan.zhihu.com/p/345214051

# 介绍
高效率、生产可用、支持快速部署的 Spark SQL Server 没有很好地解决方案。原生 Spark Thrift Server 不能很好解决多租户的问题，实现上很简单，对外提供 thrift 接口，内部通过共享 spark session 实现 spark sql 的处理，不适合在生产环境使用。

Kyuubi 提供了一个比较好的Spark SQL多租户实现方案，通过对 SQL engine 的超时回收，在性能上也有比较好的平衡。不过，在安装部署上 kyuubi 不够简单。

而 spark-sql-on-k8s 则希望提供像安装部署 mysql 或 postgresql 一样简单，来使用 spark sql。对用户提供标准的 jdbc 接口，在内则具备完整的 spark 能力，比如支持UDF、机器学习等。让 spark 用户可以只关注 sql即可，不需要具备大数据基础，也可以享受大数据技术所带来的体验。

基于 spark 实现，所以 spark-sql-on-k8s 能很容易扩展，以支持数据湖，比如 deta lake 或 iceberg 等。也能很容易地支持 ETL、BI等应用场景。快速、简单、方便、功能完成的特性，对于中小规模集群的使用体验而言，是非常值得推荐考虑的。

# 最小依赖
spark-sql-on-k8s 安装部署的依赖最小集：kubernetes、mysql(postgresql)、S3兼容存储。kubernetes 的安装部署，除了有一些快速安装工具之外，目前也有一些简化（完全兼容）实现方案（比如 k3s、k0s等），这些标准简化版方案非常适合中小规模的集群，使用上无差别，但安装部署运维上则简化许多。mysql(postgresql)的应用当然也非常广泛，基本上在任何一个集群，都有安装部署，所以对mysql的要求，实际上要做的工作非常少。首选S3兼容存储而不是HDFS，有几方面原因：一是简化运维部署难度，对于中小规模集群而言，hdfs相比S3在性价比上不占优势，同时运维也更复杂；二是拥抱云原生，现有S3兼容的云存储供应商很多，从使用成本上也有一定优势。当然，也可以选择 minio 自建S3兼容的存储层（免费但需运维支持）。

当然， 也完全支持 HDFS、hive metastore ，在标准 Hadoop 环境中使用。并且能支持 kerberos 认证方式（依赖 kyuubi 支持）。

# 开源依赖
1、spark-operator-on-k8s（spark application controller）

https://github.com/GoogleCloudPlatform/spark-on-k8s-operator

Kubernetes operator for managing the lifecycle of Apache Spark applications on Kubernetes.

2、forked kyuubi（jdbc server 和 sql engine，主要增加了 k8s SessionManager 和 K8s Session Impl 功能，以实现基于 spark operator 的 spark application 应用管理能力）。如下为开源 kyuubi 和 forked kyuubi 路径。

https://github.com/yaooqinn/kyuubi

https://github.com/yilong2001/kyuubi

Kyuubi is a high-performance universal JDBC and SQL execution engine, built on top ofApache Spark. The goal of Kyuubi is to facilitate users to handle big data like ordinary data.


# 使用步骤如下：

```
clone 代码，包含子模块。

git clone --recursive https://github.com/yilong2001/spark-sql-on-k8s.git

```

# 创建 namespace (k8s namespace)

```
kubectl create namespace spark-operator
kubectl create namespace spark-jobs
kubectl create namespace spark-history

```

# 创建 traefik v2 crds

```
kubectl apply -f traefik-helm-chart/traefik/crds/
```

# 创建 traefik v2 ingress resources。

```
首先，修改 loadbalance external ip 。
如果不设置，则使用 k8s 集群内任一 ip。

文件路径：k8s/deploy/traefik-v2/004-service.yaml

  externalIPs:
    - xxx.xxx.xxx.xxx

kubectl apply -f k8s/deploy/traefik-v2/
```

# login docker repository
```
因为 spark-operator 和 spark  镜像保存在阿里云公共镜像仓库，需要登录阿里云镜像仓库才能访问：

sudo docker login --username=yourAliyunZhangHao registry.cn-beijing.aliyuncs.com

如果是使用其他 container 工具，使用相应命令登录。

从阿里云下载镜像速度比较慢，建议镜像下载到本地仓库。
如果使用本地镜像仓库，请修改相关镜像tag：

1、registry.cn-beijing.aliyuncs.com/yilong2001/spark-operator:v1beta2-1.2.0-3.0.0

2、registry.cn-beijing.aliyuncs.com/yilong2001/spark:v3.0.1-1216

```

# 创建 spark operator on k8s 控制器 (apply spark-operator)
```
kubectl apply -f k8s/deploy/spark-operator
```

# 创建 spark sql on k8s 控制器 (apply sparkop-ctrl)
```
首先构建镜像， 需要根据 image builder (docker / img / ... ) 设置执行命名。

sh make.sh 
```

在创建控制前之前，需要修改的参数如下( 依赖 s3 和 mysql )：
```
在 k8s/deploy/sparkop-ctrl/sparkop-deployment.yml 中修改配置参数：

s3兼容的存储，可以是在公有云上对象存储；或者使用 minio 自己搭建对象存储服务。

--s3-upload-dir=s3://testbucket
--s3-endpoint=192.168.42.1:9000
--s3-accesskey=minioadmin
--s3-secretkey=minioadmin

元数据库，即 spark sql metastore 数据库，默认数据库名称为 hive ，请事先创建 hive 数据库；否则，需要配置有创建数据库权限的用户。目前只支持 mysql 数据库。

--metadb-host=192.168.42.1
--metadb-user=xxx
--metadb-pw=xxx

```

# apply spark
```
kubectl apply -f k8s/deploy/spark/
```

# apply kyuubi server
```
kubectl apply -f k8s/deploy/kyuubi-server/
```


# jdbc connect
```
jdbc 连接可以附带 spark 运行参数，以设置 memory、 cpu core 等运行参数。

如下例子：

jdbc:hive2://spark-server-jdbc-ingress:10019/?spark.dynamicAllocation.enabled=true;spark.dynamicAllocation.maxExecutors=500;spark.shuffle.service.enabled=true;spark.executor.cores=3;spark.executor.memory=2g

```

# spark ui 查看 SQL 详情
```
访问如下 URL ，可以查看某个用户下 spark sql 的历史记录 ： 
http://spark.mydomain.io/proxy/spark/spark-jobs/spark-sql-${user}

例如，使用 admin 用户登录，则访问如下 URL：
http://spark.mydomain.io/proxy/spark/spark-jobs/spark-sql-admin

```

