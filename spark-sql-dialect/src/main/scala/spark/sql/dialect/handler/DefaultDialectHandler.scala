package spark.sql.dialect.handler

import org.apache.spark.sql.catalyst.plans.logical.LogicalPlan
import org.apache.spark.sql.internal.SQLConf
import spark.sql.dialect.analysis.{DialectSqlAstBuilder, DialectSqlParser}
import spark.sql.dialect.executor.DialectCommandOperator
import spark.sql.dialect.logical.DialectRunnableCommand

class DefaultDialectHandler(conf: SQLConf, cmdOp: DialectCommandOperator) extends SqlDialectHandler {
  private val parser = new DialectSqlParser(conf)
  private val astBuilder = new DialectSqlAstBuilder(conf)

  override def sql(sqlText: String): String = {
    parser.parse(sqlText) {
      parserHandler => {
        try {
          val stat = parserHandler.singleStatement()
          val res = astBuilder.visitSingleStatement(stat)
          res match {
            case cmd: DialectRunnableCommand => {
              cmd.run(cmdOp)
            }
            //case plan: LogicalPlan => plan
            case _ => {
              throw new IllegalArgumentException("ERROR : sql not support : "+sqlText)
            }
          }
        } catch {
          case e: Exception => {
            e.printStackTrace()
            throw e
          }
        }
      }
    }
  }
}
