package spark.sql.dialect.handler

trait SqlDialectHandler {
  def sql(sqlText: String): String
}
