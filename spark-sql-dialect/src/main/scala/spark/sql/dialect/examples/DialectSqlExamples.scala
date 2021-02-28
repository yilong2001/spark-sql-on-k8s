package spark.sql.dialect.examples

import java.util

import org.apache.spark.SparkConf
import org.apache.spark.sql.SparkSession
import org.apache.spark.sql.catalyst.plans.logical.LogicalPlan
import org.apache.spark.sql.internal.SQLConf
import spark.sql.dialect.analysis.{DialectSqlAstBuilder, DialectSqlParser}
import spark.sql.dialect.executor.DialectCommandOperator
import spark.sql.dialect.handler.DefaultDialectHandler

import scala.collection.JavaConverters._

class DemoOperator extends DialectCommandOperator {
  override def createUser(name: String, password: String, sets: util.Map[String, String]): String = {
    System.out.println("create user : " + name + ":" + password)
    sets.asScala.foreach(kv => System.out.println(kv._1 + ":" + kv._2))
    "create user : " + name + " success"
  }
}

object DialectSqlExamples {
  def main(args: Array[String]): Unit = {
    val handler = new DefaultDialectHandler(new SQLConf, new DemoOperator)
    handler.sql("create user 'user1' identified by '123' settings ('key1' = 'val1', 'key2' = 'val2') ")
  }

  def createUserSql(args: Array[String]): Unit = {
    val parser = new DialectSqlParser(new SQLConf)
    val astBuilder = new DialectSqlAstBuilder(null) //sparkSession.sqlContext.conf

    parser.parse("create user 'user1' identified by '123' settings ('key1' = 'val1', 'key2' = 'val2') ") {
      parserHandler => {
        try {
          val stat = parserHandler.singleStatement()
          val res = astBuilder.visitSingleStatement(stat)
          //astBuilder.visitSingleStatement(parser.singleStatement()) match {
          res match {
            case plan: LogicalPlan => plan
            case _ => {
              throw new IllegalArgumentException()
            }
          }
        } catch {
          case e: Exception => {
            e.printStackTrace()
          }
        }
      }
    }
  }

  def main1(args: Array[String]): Unit = {
    /*val bp = "/ns/spark-jobs"
    val bpsplit = (dst : String) => {
      if (bp.length >0 && dst.startsWith(bp)) {
        dst.split(bp)(1)
      } else {
        dst
      }
    }
    val parts = Option(bpsplit("/ns/spark-jobs/history/spark-e8a6cdfb303a4210a1edb3b79475fc79/jobs/")).getOrElse("").split("/")
    System.out.println(bpsplit("/ns/spark-jobs/history/spark-e8a6cdfb303a4210a1edb3b79475fc79/jobs/"))
    System.out.println(parts.length)
    parts.foreach(System.out.println)*/

    //val list = List(Option.empty, Option("l1"), Option("l2")).filter(!_.isEmpty).map(_.get).reduce(_+","+_)
    //System.out.println(list)
    //System.exit(1)
    val sparkConf = new SparkConf
    sparkConf.setAppName("s3")
    //sparkConf.setMaster("spark://192.168.42.152:7077")
    sparkConf.setMaster("local[6]")
    sparkConf.set("spark.ui.k8sbase", "/proxy/spark-service/spark-jobs/spark-pi")
    //sparkConf.set("spark.ui.killEnabled", "true")
    //System.setProperty("spark.ui.proxyBase", "/proxy/spark-pi")
    sparkConf.set("spark.sql.parquet.mergeSchema", "false")
    sparkConf.set("spark.sql.parquet.filterPushdown", "true")
    sparkConf.set("spark.sql.hive.metastorePartitionPruning", "true")

    //sparkConf.set("spark.eventLog.enabled", "true")
    //sparkConf.set("spark.eventLog.dir", "s3a://historybucket/eventlog")//) //"file:///."
    //sparkConf.set("spark.history.fs.logDirectory", "s3a://historybucket/")
    sparkConf.set("spark.hadoop.fs.s3a.impl", "org.apache.hadoop.fs.s3a.S3AFileSystem")
    //sparkConf.set("spark.hadoop.fs.s3a.impl.disable.cache", "true")
    sparkConf.set("spark.hadoop.fs.s3a.endpoint", "192.168.42.1:9000")
    sparkConf.set("spark.hadoop.fs.s3a.access.key", "minioadmin")
    sparkConf.set("spark.hadoop.fs.s3a.secret.key", "minioadmin")
    sparkConf.set("spark.hadoop.fs.s3a.bucket.probe", "0")
    sparkConf.set("spark.hadoop.fs.s3a.connection.ssl.enabled", "false")
    sparkConf.set("spark.hadoop.fs.s3a.committer.staging.conflict-mode", "append")
    sparkConf.set("spark.hadoop.fs.hdfs.impl",
        "org.apache.hadoop.hdfs.DistributedFileSystem")
    sparkConf.set("spark.hadoop.fs.file.impl", "org.apache.hadoop.fs.LocalFileSystem")
    //sparkConf.set("spark.hadoop.fs.s3a.buffer.dir", "F:\\work\\bigdata\\code\\spark3demo\\s3spark\\tmpbuf")

    sparkConf.set("spark.hadoop.fs.s3a.aws.credentials.provider", "org.apache.hadoop.fs.s3a.TemporaryAWSCredentialsProvider, org.apache.hadoop.fs.s3a.SimpleAWSCredentialsProvider, org.apache.hadoop.fs.s3a.auth.IAMInstanceCredentialsProvider")

    /*
    sparkConf.set("spark.executor.heartbeatInterval", "6000")
    sparkConf.set("spark.storage.blockManagerSlaveTimeoutMs", "8000")
    sparkConf.set("spark.network.timeout", "8000")
    sparkConf.set("spark.network.timeoutInterval", "8000")
*/
    //sparkConf.set("spark.executor.memory", "4g")
    //sparkConf.set("spark.driver.memory", "4g")

    val sparkSession = SparkSession.builder.config(sparkConf)
      .enableHiveSupport()
      .getOrCreate

    sparkSession.sparkContext.hadoopConfiguration.set("spark.sql.parquet.mergeSchema", "false")
    sparkSession.sparkContext.hadoopConfiguration.set("spark.sql.parquet.filterPushdown", "true")
    sparkSession.sparkContext.hadoopConfiguration.set("spark.sql.hive.metastorePartitionPruning", "true")
    sparkSession.sparkContext.hadoopConfiguration.set("spark.sql.sources.commitProtocolClass", "org.apache.spark.internal.io.cloud.PathOutputCommitProtocol")
    sparkSession.sparkContext.hadoopConfiguration.set("spark.sql.parquet.output.committer.class", "org.apache.spark.internal.io.cloud.BindingParquetOutputCommitter")

    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.endpoint", "http://192.168.42.1:9000")
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.access.key", "minioadmin")
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.secret.key", "minioadmin")
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.path.style.access", "true")
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.committer.name", "directory")
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.committer.staging.abort.pending.uploads", "true")
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.committer.staging.conflict-mode", "append")
    //sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.committer.staging.tmp.path", "/tmp/staging") //: /tmp/
    //sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.committer.staging.tmp.path",
    //  "F:\\work\\bigdata\\code\\spark3demo\\s3spark\\tmp")
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.committer.staging.unique-filenames", "true")
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.connection.establish.timeout", "5000")
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.connection.ssl.enabled", "false")
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.connection.timeout", "200000")
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.impl", "org.apache.hadoop.fs.s3a.S3AFileSystem")

    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.bucket.probe", "0")

    sparkSession.sparkContext.hadoopConfiguration.set("fs.hdfs.impl",
     "org.apache.hadoop.hdfs.DistributedFileSystem");
    sparkSession.sparkContext.hadoopConfiguration.set("fs.file.impl",
      "org.apache.hadoop.fs.LocalFileSystem");

    //sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.committer.threads", "2048")// # Number of threads writing to MinIO
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.committer.threads", "2")// # Number of threads writing to MinIO
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.connection.maximum", "6")// "8192" # Maximum number of concurrent conns
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.fast.upload.active.blocks", "2048")// # Number of parallel uploads
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.fast.upload.buffer", "disk")// # Use disk as the buffer for uploads
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.fast.upload", "true")// # Turn on fast upload mode
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.max.total.tasks", "2048")// # Maximum number of parallel tasks
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.multipart.size", "512M")// # Size of each multipart chunk
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.multipart.threshold", "512M")// # Size before using multipart uploads
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.socket.recv.buffer", "65536")// # Read socket buffer hint
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.socket.send.buffer", "65536")// # Write socket buffer hint
    sparkSession.sparkContext.hadoopConfiguration.set("fs.s3a.threads.max", "3")// "2048" # Maximum number of threads for S3A

    import sparkSession.implicits._

    val sqls =
      """
        |CREATE TABLE IF NOT EXISTS  s3_parquet_t1 (
        |foo string
        |)
        |STORED AS PARQUET
        |LOCATION 's3a://testbucket/myParquet001/'
        |""".stripMargin

    val sqls2 =
      """
        |show tables;
        |""".stripMargin

    //sparkSession.sql(sqls2).show()

    //System.exit(1)
    val count = 1000//*1000
    val multi = 10  //10
    val ds = (0 to count).map(id => (id, "student_"+id)).toSeq.toDS()

    (1 to multi).foreach(x => {
      ds.map(x => { Thread.sleep(1*1000); x }).write.format("csv").save("s3a://testbucket1/ds000"+x)
      //Thread.sleep(10*1000)
    })
  }
}
