package spark.sql.dialect.executor

abstract class DialectCommandOperator {
  def createUser(name:String, password:String, sets: java.util.Map[String,String]): String
}
