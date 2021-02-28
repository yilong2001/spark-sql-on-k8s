package spark.sql.dialect.analysis

import java.util

import org.apache.log4j.Logger
import org.apache.spark.sql.catalyst.parser.ParserUtils.withOrigin
import org.apache.spark.sql.catalyst.plans.logical.LogicalPlan
import org.apache.spark.sql.internal.SQLConf
import spark.sql.dialect.executor.DialectCommandOperator
import spark.sql.dialect.logical.CreateUserCommand

import scala.collection.mutable.ArrayBuffer
import spark.sql.dialect.parser._

/**
  * Created by yilong on 2019/1/16.
  */
class DialectSqlAstBuilder(val conf : SQLConf) extends SqlBaseBaseVisitor[AnyRef] {
  val logger = Logger.getLogger(classOf[DialectSqlAstBuilder])

  override def visitKvPair(ctx: SqlBaseParser.KvPairContext): (String, String) = withOrigin(ctx) {
    (org.apache.spark.sql.catalyst.parser.ParserUtils.string(ctx.key),
      org.apache.spark.sql.catalyst.parser.ParserUtils.string(ctx.value))
  }

  override def visitSettingList(ctx: SqlBaseParser.SettingListContext): java.util.Map[String, String] = withOrigin(ctx) {
    import scala.collection.JavaConverters._

    val map = new util.HashMap[String,String]()
    ctx.kvPair().asScala.foreach(kvctx => {
      val t = visitKvPair(kvctx)
      map.put(t._1, t._2)
    })

    map
  }

  override def visitCreateUserCommand(ctx: SqlBaseParser.CreateUserCommandContext): LogicalPlan = withOrigin(ctx) {
    logger.info("start parser create user command with visitCreateUserCommand ... ")
    val name:String = org.apache.spark.sql.catalyst.parser.ParserUtils.string(ctx.name)
    val password:String = org.apache.spark.sql.catalyst.parser.ParserUtils.string(ctx.password)
    val sets = if (ctx.sets != null) {
      visitSettingList(ctx.sets)
    } else {
      new java.util.HashMap[String,String]
    }
    logger.info(name + ":" + password)
    logger.info(sets)

    CreateUserCommand(name, password, sets)
  }

  override def visitSingleStatement(ctx: SqlBaseParser.SingleStatementContext): LogicalPlan = withOrigin(ctx) {
    visit(ctx.statement).asInstanceOf[LogicalPlan]
  }
}
