# spark-sql-on-k8s
最简单的 spark sql on kubernetes 生产环境部署方案

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


